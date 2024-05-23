package server

type Option func(s *HTTPServer)

func WithTplEngine(eg TemplateEngine) Option {
	return func(s *HTTPServer) {
		s.tplEngine = eg
	}
}
