package rest

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type {{.Struct.UnexportedName}} struct {
    {{range $f := .Struct.Fields}}{{if $f.IsID}}idType{{else}}{{$f.Name}} {{$f.RestType}} `{{$f.RestTag}}`{{end}}
    {{end}}
}

func from{{.Struct.ExportedName}}(in domain.{{.Struct.ExportedName}}) (*{{.Struct.UnexportedName}}, error) {
	out := {{.Struct.UnexportedName}}{
        {{with $s := .Struct}}{{range $f := $s.Fields}}{{$f.RestConvert $s}}
    {{end}}{{end}}
	}
    return &out, nil
}


func from{{.Struct.ExportedPluralName}}(ins domain.{{.Struct.ExportedPluralName}}) ([]{{.Struct.UnexportedName}}, error) {
	out := make([]{{.Struct.UnexportedName}}, len(ins))
	for i, in := range ins {
		v, err := from{{.Struct.ExportedName}}(in)
		if err != nil {
			return nil, err
		}
		out[i] = *v
	}
	return out, nil
}

type req{{.Struct.ExportedName}} struct {
}

func (q req{{.Struct.ExportedName}}) to() (*domain.{{.Struct.ExportedName}}, error) {
	out := domain.{{.Struct.ExportedName}}{}
    return &out, nil
}

type resp{{.Struct.ExportedName}} struct {
	{{.Struct.ExportedPluralName}} []{{.Struct.UnexportedName}} `json:"{{.Struct.PluralJSONName}}"`
}

func buildResp{{.Struct.ExportedName}}({{.Struct.UnexportedPluralName}} domain.{{.Struct.ExportedPluralName}}) (*resp{{.Struct.ExportedName}}, error) {
	resp := resp{{.Struct.ExportedName}}{}
	var err error
	resp.{{.Struct.ExportedPluralName}}, err = from{{.Struct.ExportedPluralName}}({{.Struct.UnexportedPluralName}})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create{{.Struct.ExportedName}} godoc
// @Summary Create {{.Struct.UnexportedName}}
// @Tags {{.Struct.ExportedName}}
// @Param questionID path string true " "
// @Param with query []string false " "
// @Param body body req{{.Struct.ExportedName}} true " "
// @Success 201 {object} envelope{data=resp{{.Struct.ExportedName}}} ""
// @Router /users/{userID}/lessons/{lessonID}/questions/{questionID}/answers  [post]
func (h *handler) Create{{.Struct.ExportedPluralName}}(ctx *gin.Context) {
	var req req{{.Struct.ExportedName}}
	if err := ctx.Bind(&req); err != nil {
		h.raiseError(ctx, err)
		return
	}
	// userID := ctx.Param("userID")
	in, err := req.to()
	if err != nil {
		h.raiseError(ctx, err)
		return
	}
	with := domain.NewWith(ctx.QueryArray("with"))
	out, err := h.backend.Create{{.Struct.ExportedName}}(context.Background(), *in, with)
	if err != nil {
		h.raiseError(ctx, err)
		return
	}
	resp, err := buildResp{{.Struct.ExportedName}}(domain.{{.Struct.ExportedPluralName}}{*out.{{.Struct.ExportedName}}})
	if err != nil {
		h.raiseError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, envelope{Data: resp})
}