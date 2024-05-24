package session

import "jungle/session"

type Option func(m *Middleware)

func WithSessionManager(managerCreate func() *session.Manager) Option {
	return func(m *Middleware) {
		manager := managerCreate()
		m.manager = manager
	}
}
