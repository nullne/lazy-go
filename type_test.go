package main_test

import (
	"os"
	"testing"
	"text/template"

	goes "github.com/nullne/lazy-go"
)

func TestPlural(t2 *testing.T) {
	tpl := template.Must(template.New("letter").ParseFiles("/Users/nullne/go/src/github.com/nullne/layz-go/templates/db.tpl"))

	r := goes.Foo{
		GoModule: "abc.com/nice/you",
		Struct: goes.Struct{
			Name: "GoSwimming",
			Fields: goes.Fields{
				{"ID", "string"},
				{"UserID", "string"},
				{"TimeSlots", "[]string"},
				{"Duration", "int"},
				{"CreatedAt", "time.Time"},
			},
		},
	}
	err := tpl.ExecuteTemplate(os.Stdout, "db.tpl", r)
	if err != nil {
		panic(err)
	}
	// p := pluralize.NewClient()
	// fmt.Println(p.Plural("teacher"))
}
