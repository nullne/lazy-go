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
	out := make([]string, 0, len(fs))

	for _, f := range fs {
		if f.IsID() {
			continue
		}
		out = append(out, fmt.Sprintf(":%s", f.Name.SnakeCase()))
	}
	return strings.Join(out, ", ")
}

type Field struct {
	Name        VarName
	Type        string
	LocalStruct bool
	IsEnum      bool
}

func (f Field) IsID() bool {
	return f.Name == "ID"
}

func (f Field) IsIDType() bool {
	return strings.HasSuffix(string(f.Name), "ID") ||
		strings.HasSuffix(string(f.Name), "IDs")
}

const (
	ConvertDBFrom1 = iota
	ConvertDBFrom2
	ConvertDBTo1
	ConvertDBTo2
	ConvertDBType
	ConvertDBTag
)

func (f Field) ConvertDBFrom1() string {
	return f.convert(ConvertDBFrom1)
}

func (f Field) ConvertDBFrom2() string {
	return f.convert(ConvertDBFrom2)
}

func (f Field) ConvertDBTo1() string {
	return f.convert(ConvertDBTo1)
}

func (f Field) ConvertDBTo2() string {
	return f.convert(ConvertDBTo2)
}

func (f Field) ConvertDBType() string {
	return f.convert(ConvertDBType)
}

func (f Field) ConvertDBTag() string {
	return f.convert(ConvertDBTag)
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

// func (f Field) DBConvertTo() string {
// 	switch f.Type {
// 	case "*time.Time":
// 		return ""
// 	case "time.Time":
// 		return fmt.Sprintf(`%s: in.%s.UTC(),`, f.Name, f.Name)
// 	case "[]int":
// 		return fmt.Sprintf(`%s: fromInt64Array(in.%s),`, f.Name, f.Name)
// 	default:
// 		return fmt.Sprintf(`%s: in.%s,`, f.Name, f.Name)
// 	}
// }

// func (f Field) DBConvertToComplicated() string {
// 	switch f.Type {
// 	case "*time.Time":
// 		tpl := `if in.%s.Valid {
// 			out.%s = &in.%s.Time.UTC()
// 		}`
// 		return fmt.Sprintf(tpl, f.Name, f.Name, f.Name)

// 	case "*string":
// 		tpl := `if in.%s.Valid {
// 			out.%s = &in.%s.String
// 		}`
// 		return fmt.Sprintf(tpl, f.Name, f.Name, f.Name)
// 	default:
// 		return ""
// 	}
// }

// func (f Field) DBConvertFrom() string {
// 	switch f.Type {
// 	case "time.Time":
// 		return fmt.Sprintf(`%s: in.%s.UTC(),`, f.Name, f.Name)
// 	case "[]int":
// 		return fmt.Sprintf(`%s: toInt64Array(in.%s),`, f.Name, f.Name)
// 	default:
// 		return fmt.Sprintf(`%s: in.%s,`, f.Name, f.Name)
// 	}
// }

func (f Field) convert(convertType int) string {
	switch f.Type {
	case "time.Time":
		return f.convertTime(convertType)
	case "*time.Time":
		return f.convertNullTime(convertType)
	case "[]string":
		return f.convertStringArray(convertType)
	case "[]int":
		return f.convertIntArray(convertType)
	case "*string":
		return f.convertNullString(convertType)
	default:
		return f.convertDefault(convertType)
	}
}

// one one
func (f Field) convertDefault(convertType int) string {
	switch convertType {
	case ConvertDBType:
		if f.IsEnum {
			return "string"
		} else if f.LocalStruct {
			return "[]byte"
		} else {
			return f.Type
		}
	case ConvertDBFrom1:
		if f.IsEnum {
			return fmt.Sprintf(`%s: in.%s.String(),`, f.Name, f.Name)
		} else if f.LocalStruct {
			return ""
		} else {
			return fmt.Sprintf(`%s: in.%s,`, f.Name, f.Name)
		}
	case ConvertDBFrom2:
		if f.IsEnum {
			return ""
		} else if f.LocalStruct {
			tpl := `%s, err := json.Marshal(in.%s)
			if err != nil {
				return nil, err
			}
			out.%s = %s
			`
			return fmt.Sprintf(tpl, f.Name.LowerFirstLetter(), f.Name, f.Name, f.Name.LowerFirstLetter())
		} else {
			return ""
		}
	case ConvertDBTo1:
		if f.IsEnum {
			return ""
		} else if f.LocalStruct {
			return ""
		} else {
			return fmt.Sprintf(`%s: in.%s,`, f.Name, f.Name)
		}
	case ConvertDBTo2:
		if f.IsEnum {
			tpl := `%s, err := domain.%sString(in.%s)
			if err != nil {
				return nil, err
			}
			out.%s = %s
			`
			return fmt.Sprintf(tpl, f.Name.LowerFirstLetter(), f.Type, f.Name, f.Name, f.Name.LowerFirstLetter())

		} else if f.LocalStruct {
			tpl := `var %s domain.%s
			if err := json.Unmarshal(in.%s, &%s); err != nil {
				return nil, err
			}
			out.%s = %s
			`
			return fmt.Sprintf(tpl, f.Name.LowerFirstLetter(), f.Type, f.Name, f.Name.LowerFirstLetter(), f.Name, f.Name.LowerFirstLetter())

		} else {
			return ""
		}
	default:
		panic("unknown convertType")
	}
}

func (f Field) convertTime(convertType int) string {
	switch convertType {
	case ConvertDBType:
		return f.Type
	case ConvertDBFrom1:
		return fmt.Sprintf(`%s: in.%s.UTC(),`, f.Name, f.Name)
	case ConvertDBFrom2:
		return ""
	case ConvertDBTo1:
		return fmt.Sprintf(`%s: in.%s.UTC(),`, f.Name, f.Name)
	case ConvertDBTo2:
		return ""
	default:
		panic("unknown convertType")
	}
}

func (f Field) convertNullTime(convertType int) string {
	switch convertType {
	case ConvertDBType:
		return "sql.NullTime"
	case ConvertDBFrom1:
		return ""
	case ConvertDBFrom2:
		tpl := `if in.%s != nil {
			if err := out.%s.Scan(in.%s.UTC()); err != nil {
				return nil, err
			}
		}`
		return fmt.Sprintf(tpl, f.Name, f.Name, f.Name)
	case ConvertDBTo1:
		return ""
	case ConvertDBTo2:
		tpl := `if in.%s.Valid {
			v := in.%s.Time.UTC()
			out.%s = &v
		}`
		return fmt.Sprintf(tpl, f.Name, f.Name, f.Name)
	default:
		panic("unknown convertType")
	}
}

func (f Field) convertStringArray(convertType int) string {
	switch convertType {
	case ConvertDBType:
		return "pq.StringArray"
	case ConvertDBFrom1:
		return f.convertDefault(convertType)
	case ConvertDBFrom2:
		return ""
	case ConvertDBTo1:
		return f.convertDefault(convertType)
	case ConvertDBTo2:
		return ""
	default:
		panic("unknown convertType")
	}
}

func (f Field) convertIntArray(convertType int) string {
	switch convertType {
	case ConvertDBType:
		return "pq.Int64Array"
	case ConvertDBFrom1:
		return fmt.Sprintf(`%s: toInt64Array(in.%s),`, f.Name, f.Name)
	case ConvertDBFrom2:
		return ""
	case ConvertDBTo1:
		return fmt.Sprintf(`%s: fromInt64Array(in.%s),`, f.Name, f.Name)
	case ConvertDBTo2:
		return ""
	default:
		panic("unknown convertType")
	}
}

func (f Field) convertNullString(convertType int) string {
	switch convertType {
	case ConvertDBType:
		return "sql.NullString"
	case ConvertDBFrom1:
		return ""
	case ConvertDBFrom2:
		tpl := `if in.%s != nil {
			if err := out.%s.Scan(*in.%s); err != nil {
				return nil, err
			}
		}`
		return fmt.Sprintf(tpl, f.Name, f.Name, f.Name)
	case ConvertDBTo1:
		return ""
	case ConvertDBTo2:
		tpl := `if in.%s.Valid {
			out.%s = &in.%s.String
		}`
		return fmt.Sprintf(tpl, f.Name, f.Name, f.Name)
	default:
		panic("unknown convertType")
	}
}

// func (f Field) convert(convertType int) string {
// 	switch convertType {
// 	case convertTypeType:
// 		return f.Type
// 	case convertTypeFrom1:
// 		return ""
// 	case convertTypeFrom2:
// 		return ""
// 	case convertTypeTo1:
// 		return ""
// 	case convertTypeTo2:
// 		return ""
// 	default:
// 		return ""
// 	}
// }

// func (f Field) DBConvertFromComplicated() string {
// 	switch f.Type {
// 	case "time.Time":
// 		return fmt.Sprintf(`%s: in.%s.UTC(),`, f.Name, f.Name)
// 	case "[]int":
// 		return fmt.Sprintf(`%s: toInt64Array(in.%s),`, f.Name, f.Name)
// 	default:
// 		return fmt.Sprintf(`%s: in.%s,`, f.Name, f.Name)
// 	}
// }

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
	s := strings.Join(lowerSplitted, "_")
	s = strings.ReplaceAll(s, "_i_ds", "_ids")
	return VarName(s)
}
