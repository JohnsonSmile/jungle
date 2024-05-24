package cookie

import (
	"net/http"
)

type Propagator struct {
	CookieName string
	Domain     string
	CookieOpts []func(c *http.Cookie)
}

func NewPropagator(cookieName string, domain string, opts ...Option) *Propagator {
	p := &Propagator{
		CookieName: cookieName,
		Domain:     domain,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Propagator) Inject(sessionId string, writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:     p.CookieName,
		Value:    sessionId,
		Domain:   p.Domain,
		HttpOnly: true,
	}
	for _, opt := range p.CookieOpts {
		opt(cookie)
	}
	http.SetCookie(writer, cookie)
	return nil
}

func (p *Propagator) Extract(req *http.Request) (sessionId string, err error) {
	cookie, err := req.Cookie(p.CookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, err
}

func (p *Propagator) Remove(writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:   p.CookieName,
		MaxAge: -1,
	}
	http.SetCookie(writer, cookie)
	return nil
}
