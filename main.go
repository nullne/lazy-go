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
	"text/template"
)

type foo struct {
	AnotherPeople string `db:"another_people" json:"another_people,omitempty" yaml:"another_people"`
	B             string `db:"b" json:"b,omitempty" yaml:"b"`
	C             struct {
		Hi         string `json:"hi,omitempty"`
		YouAreSHIT string `json:"you_are_shit,omitempty"`
	} `json:"c,omitempty"`
}

var (
	gFile = flag.String("file", "./testdata/hello.go", "specify go file")
	gLine = flag.Int("line", -1, "line which the target struct contains")
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

	r := Foo{
		Struct: Struct{},
	}

	structNames := collectStructs(file)
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
				fields = append(fields, Field{Name: VarName(name.Name), Type: typeNameBuf.String()})
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

	tpl := template.Must(template.New("letter").ParseFiles("/Users/nullne/go/src/github.com/nullne/layz-go/templates/db.tpl"))

	err = tpl.ExecuteTemplate(os.Stdout, "db.tpl", r)
	if err != nil {
		return err
	}
	return nil
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
