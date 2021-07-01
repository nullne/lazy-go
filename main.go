package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

var (
	gFile = flag.String("file", "./testdata/hello.go", "specify go file")
	gLine = flag.Int("line", -1, "line which the target struct contains")
	gType = flag.String("type", "pg", "what kind of convert, support pg, rest")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	src, err := ioutil.ReadFile(*gFile)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "demo", src, parser.ParseComments)
	if err != nil {
		return err
	}

	// why comments
	comments := collectComments(file.Comments, fset)
	// fmt.Println(comments)

	r := Foo{
		Struct: Struct{},
	}

	structNames := collectStructs(file)
	types := collectTypes(file)
	// fmt.Println(types)

	var fields Fields
	ast.Inspect(file, func(x ast.Node) bool {

		s, ok := x.(*ast.StructType)
		if !ok {
			return true
		}
		startLine := fset.Position(x.Pos()).Line
		endLine := fset.Position(x.End()).Line

		// fmt.Printf("(%d - %d) find %d\n", startLine, endLine, *gLine)
		if startLine > *gLine || endLine < *gLine {
			return true
		}
		// fmt.Printf("new struct from %d to %d.(%d-%d)\n", s.Struct, s.End(), s.Fields.Opening, s.Fields.Closing)

		for _, field := range s.Fields.List {
			// v := reflect.TypeOf(field.Type)
			var typeNameBuf bytes.Buffer
			printer.Fprint(&typeNameBuf, fset, field.Type)
			// fmt.Println("type", field.Type, "---", v.String(), "----", typeNameBuf.String())
			for _, name := range field.Names {
				fields = append(fields, Field{
					Name:        VarName(name.Name),
					Type:        typeNameBuf.String(),
					LocalStruct: localStruct(structNames, typeNameBuf.String()),
					IsEnum:      hasString(comments, findStructLine(fset, types, typeNameBuf.String())),
				})
				// fmt.Printf("%d %s %#v\n", name.NamePos, name.Name, *(name.Obj))
			}
			// fmt.Printf("Field: %s\n", field.Names[0].Name)
			// fmt.Printf("Tag:   %s\n", field.Tag.Value)
		}

		r.Struct.Name = VarName(structNames[x.Pos()])
		return false
	})
	if len(fields) == 0 {
		return errors.New("please select the struct to generate db implementation")
	}
	r.Struct.Fields = fields
	// fmt.Println(fields)

	var tplPath string
	switch *gType {
	case "pg":
		tplPath = "/Users/nullne/go/src/github.com/nullne/layz-go/templates/db.tpl"
	case "rest":
		tplPath = "/Users/nullne/go/src/github.com/nullne/layz-go/templates/rest.tpl"
	}
	tpl := template.Must(template.New("letter").ParseFiles(tplPath))

	err = tpl.ExecuteTemplate(os.Stdout, path.Base(tplPath), r)
	if err != nil {
		return err
	}
	return nil
}

func localStruct(s map[token.Pos]string, name string) bool {
	for _, v := range s {
		if v == name {
			return true
		}
	}
	return false
}

func hasString(comments map[int]string, line int) bool {
	if line == -1 {
		return false
	}
	s, ok := comments[line-1]
	if !ok {
		return false
	}
	if !strings.Contains(s, "go:generate") {
		return false
	}
	if !strings.Contains(s, "enumer") {
		return false
	}
	return true
}

func findStructLine(fset *token.FileSet, structs map[token.Pos]string, name string) int {
	var line int = -1
	for p, n := range structs {
		if n == name {
			line = fset.Position(p).Line
			return line
		}
	}
	return line
}

func collectComments(comments []*ast.CommentGroup, fset *token.FileSet) map[int]string {
	docs := make(map[int]string)
	for _, c := range comments {
		line := fset.Position(c.Pos()).Line
		docs[line] = c.Text()
	}
	return docs
}

func collectStructs(node ast.Node) map[token.Pos]string {
	structs := make(map[token.Pos]string, 0)
	collectStructs := func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Type == nil {
			return true
		}

		structName := t.Name.Name

		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structs[x.Pos()] = structName
		return true
	}
	ast.Inspect(node, collectStructs)
	return structs
}

func collectTypes(node ast.Node) map[token.Pos]string {
	structs := make(map[token.Pos]string, 0)
	collectStructs := func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structs[t.Pos()] = t.Name.Name
		return true
	}
	ast.Inspect(node, collectStructs)
	return structs
}
