package cookie

import "net/http"

type Option func(p *Propagator)

func WithCookieOptions(opts ...func(cookie *http.Cookie)) Option {
	return func(p *Propagator) {
		p.CookieOpts = opts
	}
}
