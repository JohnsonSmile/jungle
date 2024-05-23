package engines

import (
	"bytes"
	"context"
	"html/template"
)

type GoTemplateEngine struct {
	T *template.Template
}

func (t *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := t.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}
