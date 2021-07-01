package db

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type {{.Struct.Name.LowerFirstLetter}} struct {
    {{range .Struct.Fields}}{{.Name}} {{.ConvertDBType}} `{{.DBTag}}`
    {{end }}}

func (in {{.Struct.Name.LowerFirstLetter}}) to() (*domain.{{.Struct.Name}}, error) {
	out := domain.{{.Struct.Name}}{
        {{range .Struct.Fields}}{{with $v := .ConvertDBTo1}}{{$v}}
    {{end}}{{end }}}
	{{range .Struct.Fields}}{{with $v := .ConvertDBTo2}}{{$v}}
	{{end}}{{end}}
	return &out, nil
}

type {{.Struct.Name.LowerFirstLetter.Plural}} []{{.Struct.Name.LowerFirstLetter}}

func (ins {{.Struct.Name.LowerFirstLetter.Plural}}) to() ([]domain.{{.Struct.Name}}, error) {
	outs := make([]domain.{{.Struct.Name}}, len(ins))
	for i, in := range ins {
		v, err := in.to()
		if err != nil {
			return nil, err
		}
		outs[i] = *v
	}
	return outs, nil
}

func from{{.Struct.Name}}(in domain.{{.Struct.Name}}) (*{{.Struct.Name.LowerFirstLetter}}, error) {
	out := {{.Struct.Name.LowerFirstLetter}}{
		{{range .Struct.Fields}}{{with $v := .ConvertDBFrom1}}{{$v}}
	{{end}}{{end}}}
	{{range .Struct.Fields}}{{with $v := .ConvertDBFrom2}}{{$v}}
	{{end}}{{end}}
	return &out, nil
}

func from{{.Struct.Name.Plural}}(ins []domain.{{.Struct.Name}}) ([]{{.Struct.Name.LowerFirstLetter}}, error) {
	outs := make([]{{.Struct.Name.LowerFirstLetter}}, len(ins))
	for i, in := range ins {
		v, err := from{{.Struct.Name}}(in)
		if err != nil {
			return nil, err
		}
		outs[i] = *v
	}
	return outs, nil
}

func (m *Manager) Insert{{.Struct.Name}}(ctx context.Context, in domain.{{.Struct.Name}}) (*domain.{{.Struct.Name}}, error) {
	query, args, err := m.build{{.Struct.Name.Plural}}Query(ctx, in)
	if err != nil {
		return nil, err
	}
	var out {{.Struct.Name.LowerFirstLetter}}
	if err := m.core.GetContextOnMaster(ctx, &out, query, args...); err != nil {
		return nil, err
	}
	return out.to()
}

func (m *Manager) Insert{{.Struct.Name.Plural}}(ctx context.Context, ins domain.{{.Struct.Name.Plural}}) (domain.{{.Struct.Name.Plural}}, error) {
	query, args, err := m.build{{.Struct.Name.Plural}}Query(ctx, ins...)
	if err != nil {
		return nil, err
	}
	var out {{.Struct.Name.LowerFirstLetter.Plural}}
	if err := m.core.SelectContextOnMaster(ctx, &out, query, args...); err != nil {
		return nil, err
	}
	return out.to()
}

func (m *Manager) build{{.Struct.Name.Plural}}Query(ctx context.Context, ins ...domain.{{.Struct.Name}}) (string, []interface{}, error) {
	s := "INSERT INTO {{.Struct.Name.SnakeCase.Plural}} ({{.Struct.Fields.Columns}}) VALUES ({{.Struct.Fields.NamedBindVars}}) RETURNING *"
	data, err := from{{.Struct.Name.Plural}}(ins)
	if err != nil {
		return "", nil, err
	}
	query, args, err := sqlx.Named(s, data)
	if err != nil {
		return "", nil, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	return query, args, nil
}

func (m *Manager) Select{{.Struct.Name}}ByID(ctx context.Context, id string) (*domain.{{.Struct.Name}}, error) {
	{{.Struct.Name.LowerFirstLetter.Plural}}, err := m.Select{{.Struct.Name.Plural}}ByIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	if len({{.Struct.Name.LowerFirstLetter.Plural}}) != 1 {
		return nil, sql.ErrNoRows
	}
	return &({{.Struct.Name.LowerFirstLetter.Plural}}[0]), nil
}

func (m *Manager) Select{{.Struct.Name.Plural}}ByIDs(ctx context.Context, ids ...string) (domain.{{.Struct.Name.Plural}}, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	s := "SELECT * FROM {{.Struct.Name.SnakeCase.Plural}} WHERE id IN (?)"
	query, args, err := sqlx.In(s, ids)
	if err != nil {
		return nil, err
	}
	// must rebind with $ for postgres
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	var data {{.Struct.Name.LowerFirstLetter.Plural}}
	if err := m.core.SelectContext(ctx, &data, query, args...); err != nil {
		return nil, err
	}
	return data.to()
}