//go:build integration

package engines

import (
	"html/template"
	"jungle/server"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test --tags=integration -v engines/*.go -run TestGoTemplateEngine

func TestGoTemplateEngine(t *testing.T) {
	_, filePath, _, _ := runtime.Caller(0)
	tplPath := path.Join(path.Dir(filePath), "testdata", "tpls", "*.gohtml")
	tpl := template.New("default")
	tpl, err := tpl.ParseGlob(tplPath)
	require.NoError(t, err)
	eg := &GoTemplateEngine{
		T: tpl,
	}

	serv := server.New(":8081", server.WithTplEngine(eg))
	serv.Get("/login", func(ctx *server.Context) {
		ctx.Render("login", nil)
	})
	err = serv.Start()
	require.NoError(t, err)
	err = serv.ShutDown()
	require.NoError(t, err)
}
