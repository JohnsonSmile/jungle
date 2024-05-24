package session

import (
	"jungle/server"
)

type Manager struct {
	p          Propagator
	s          Store
	sessionKey string
	idGen      func(ctx *server.Context) string
}

func NewManager(p Propagator, s Store, sessionKey string, idGen func(ctx *server.Context) string) *Manager {
	return &Manager{
		p:          p,
		s:          s,
		sessionKey: sessionKey,
		idGen:      idGen,
	}
}

func (m *Manager) GetSession(ctx *server.Context) (Session, error) {

	// 尝试从 context 中获取 session
	sessValue, exists := ctx.Get(m.sessionKey)
	sess, ok := sessValue.(Session)
	if exists && ok {
		return sess, nil
	}

	sessionId, err := m.p.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	sess, err = m.s.Get(ctx.Req.Context(), sessionId)
	if err != nil {
		return nil, err
	}
	// 将 session 写入到
	ctx.Set(m.sessionKey, sess)
	return sess, nil
}

func (m *Manager) GenerateSession(ctx *server.Context) (Session, error) {
	id := m.idGen(ctx)
	sess, err := m.s.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	// 注入到 response header 中
	err = m.p.Inject(id, ctx.Resp)
	return sess, err
}

func (m *Manager) RefreshSession(ctx *server.Context) error {
	session, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	return m.s.Refresh(ctx.Req.Context(), session.ID())
}

func (m *Manager) RemoveSession(ctx *server.Context) error {

	sessionId, err := m.p.Extract(ctx.Req)
	if err != nil {
		return err
	}
	err = m.s.Remove(ctx.Req.Context(), sessionId)
	if err != nil {
		return err
	}
	ctx.Del(m.sessionKey)
	return m.p.Remove(ctx.Resp)
}
