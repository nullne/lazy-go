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

func (s Struct) UnexportedName() string {
	return string(s.Name.LowerFirstLetter())
}

func (s Struct) ExportedName() string {
	return string(s.Name)
}

func (s Struct) UnexportedPluralName() string {
	return string(s.Name.LowerFirstLetter().Plural())
}

func (s Struct) ExportedPluralName() string {
	return string(s.Name.Plural())
}

func (s Struct) PluralJSONName() string {
	return string(s.Name.CamelCase().Plural().LowerFirstLetter())
}

func (s Struct) TableName() string {
	return string(s.Name.SnakeCase().Plural())
}

var (
	pc = pluralize.NewClient()
)

type Fields []Field

func (fs Fields) Columns() string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		if f.IsID() {
			continue
		}
		out = append(out, string(f.Name.SnakeCase()))
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

func (f Field) IsID() bool {
	return f.Name == "ID"
}

func (f Field) IsIDType() bool {
	return strings.HasSuffix(string(f.Name), "ID") ||
		strings.HasSuffix(string(f.Name), "IDs")
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
func (f Field) RestType() string {
	switch f.Type {
	case "time.Time":
		return "Iso8601Time"
	case "string":
		if f.IsIDType() {
			return "idType"
		} else {
			return f.Type
		}
	case "[]string":
		if f.IsIDType() {
			return "[]idType"
		} else {
			return f.Type
		}

	default:
		return f.Type
	}
}

func (f Field) RestConvert(s Struct) string {
	if f.IsID() {
		return fmt.Sprintf(`idType: it%s(in.ID),`, s.ExportedName())
	} else if f.IsIDType() {
		switch f.Type {
		case "string":
			return fmt.Sprintf(`%s: it%s(in.%s),`, f.Name, f.Name.ResourceName(), f.Name)
		case "[]string":
			return fmt.Sprintf(`%s: make([]idType, len(in.%s),`, f.Name, f.Name)
		default:
			panic("unsupported type")
		}
	} else if f.Type == "time.Time" {
		return fmt.Sprintf(`%s: *toUtcIsoTime(in.%s),`, f.Name, f.Name)
	} else {
		return fmt.Sprintf(`%s: in.%s,`, f.Name, f.Name)

	}
}

func (f Field) DBTag() string {
	return fmt.Sprintf(`db:"%s"`, f.Name.SnakeCase())
}

func (f Field) RestTag() string {
	s := `json:"%s"`
	if f.Type == "time.Time" {
		s = `json:"%s"  swaggertype:"primitive,string"`
	}
	return fmt.Sprintf(s, f.Name.CamelCase().LowerFirstLetter())
}

type VarName string

func (n VarName) ResourceName() string {
	if strings.HasSuffix(string(n), "ID") {
		return strings.TrimSuffix(string(n), "ID")
	} else if strings.HasSuffix(string(n), "IDs") {
		return strings.TrimSuffix(string(n), "IDs")
	}
	return string(n)
}

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
