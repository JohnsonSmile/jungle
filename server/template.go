package server

import "context"

type TemplateEngine interface {
	// Reander 渲染页面
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}
