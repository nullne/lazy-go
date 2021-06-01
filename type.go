package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
	"github.com/gertd/go-pluralize"
)

type Foo struct {
	GoModule string
	Struct   Struct
}

type Struct struct {
	Name   VarName
	Fields Fields
}

func (s Struct) TableName() string {
	return ""
}

var (
	pc = pluralize.NewClient()
)

type Fields []Field

func (fs Fields) Columns() string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = string(f.Name.SnakeCase())
	}
	return strings.Join(out, ", ")
}

func (fs Fields) NamedBindVars() string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = fmt.Sprintf(":%s", f.Name.SnakeCase())
	}
	return strings.Join(out, ", ")
}

type Field struct {
	Name VarName
	Type string
}

func (f Field) DBType() string {
	switch f.Type {
	case "[]string":
		return "pq.StringArray"
	case "[]int":
		return "pq.Int64Array"
	default:
		return f.Type
	}
}

func (f Field) DBTag() string {
	return fmt.Sprintf(`db:"%s"`, f.Name.SnakeCase())
}

func (f Field) JSONTag() string {
	return fmt.Sprintf(`json:"%s"`, f.Name.CamelCase())
}

type VarName string

func (n VarName) LowerFirstLetter() VarName {
	for i, v := range n {
		return VarName(unicode.ToLower(v)) + n[i+1:]
	}
	return ""
}

func (n VarName) Plural() VarName {
	return VarName(pc.Plural(string(n)))
}

// Must be invoked at the very begining
func (n VarName) CamelCase() VarName {
	splitted := camelcase.Split(string(n))
	var titled []string
	for _, s := range splitted {
		titled = append(titled, strings.Title(s))
	}

	titled[0] = strings.ToLower(titled[0])

	return VarName(strings.Join(titled, ""))
}

// Must be invoked at the very begining
func (n VarName) SnakeCase() VarName {
	splitted := camelcase.Split(string(n))
	var lowerSplitted []string
	for _, s := range splitted {
		lowerSplitted = append(lowerSplitted, strings.ToLower(s))
	}

	return VarName(strings.Join(lowerSplitted, "_"))
}
