package server

import (
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
)

type Uploader struct {
}

func (u *Uploader) Handle(filename string, dstPathFunc func(header *multipart.FileHeader) string) HandleFunc {
	return func(ctx *Context) {
		file, header, err := ctx.FormFile(filename)
		if err != nil {
			ctx.WriteString(http.StatusInternalServerError, []byte("上传失败:读取文件流失败:"+err.Error()))
			return
		}
		defer file.Close()
		// 目标路径
		dstPath := dstPathFunc(header)
		dstDir := path.Dir(dstPath)
		_, err = os.Stat(dstDir)
		if err != nil && errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(dstDir, 0777)
			if err != nil {
				ctx.WriteString(http.StatusInternalServerError, []byte("上传失败:创建路径失败:"+err.Error()))
				return
			}
		}

		// 目标文件
		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			ctx.WriteString(http.StatusInternalServerError, []byte("上传失败:创建文件失败:"+err.Error()))
			return
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, file)
		if err != nil {
			ctx.WriteString(http.StatusInternalServerError, []byte("上传失败:保存文件失败:"+err.Error()))
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"code": http.StatusOK,
			"msg":  "success",
		})
	}
}

type Downloader struct {
}

func (d *Downloader) Handle(dir string) HandleFunc {
	return func(ctx *Context) {
		result := ctx.QueryValue("file")
		if result.err != nil {
			ctx.WriteString(http.StatusBadRequest, []byte("params file could not found"))
			return
		}
		// path.Clean 避免别人通过../../作为文件名,来下载其他文件,防攻击
		dst := path.Join(dir, path.Clean(result.val))
		if !strings.Contains(dst, dir) {
			ctx.WriteString(http.StatusBadRequest, []byte("params file should be in dir"))
			return
		}
		filename := path.Base(dst)
		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+filename)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		// TODO: 缓存问题...
		http.ServeFile(ctx.Resp, ctx.Req, dst)
	}
}
