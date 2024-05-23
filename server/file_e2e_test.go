//go:build integration

package server

import (
	"html/template"
	"jungle/engines"
	"mime/multipart"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test --tags=integration -v server/*.go -run TestFileUpload

func TestFileUpload(t *testing.T) {
	_, filePath, _, _ := runtime.Caller(0)
	testdataDir := path.Join(path.Dir(filePath), "testdata")
	tplPath := path.Join(testdataDir, "tpls", "*.gohtml")
	tpl := template.New("default")
	tpl, err := tpl.ParseGlob(tplPath)
	require.NoError(t, err)
	eg := &engines.GoTemplateEngine{
		T: tpl,
	}

	serv := New(":8081", WithTplEngine(eg))
	serv.Get("/upload", func(ctx *Context) {
		ctx.Render("upload", nil)
	})
	uploader := &Uploader{}
	serv.Post("/upload", uploader.Handle("myfile", func(header *multipart.FileHeader) string {
		return path.Join(testdataDir, "files", header.Filename)
	}))
	err = serv.Start()
	require.NoError(t, err)
	err = serv.ShutDown()
	require.NoError(t, err)
}

// go test --tags=integration -v server/*.go -run TestFileDownload

func TestFileDownload(t *testing.T) {
	_, filePath, _, _ := runtime.Caller(0)
	testdataDir := path.Join(path.Dir(filePath), "testdata", "files")

	serv := New(":8081")
	downloader := &Downloader{}
	serv.Get("/download", downloader.Handle(testdataDir))
	err := serv.Start()
	require.NoError(t, err)
	err = serv.ShutDown()
	require.NoError(t, err)
}
