package memory

import (
	"context"
	"errors"
	"github.com/patrickmn/go-cache"
	"jungle/session"
	"sync"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not exists")
)

type Store struct {
	mux        sync.RWMutex
	sessions   *cache.Cache
	expiration time.Duration
}

func NewStore(expiration time.Duration, expireCheckInterval time.Duration) *Store {
	return &Store{
		sessions:   cache.New(expiration, expireCheckInterval),
		expiration: expiration,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mux.Lock()
	s.mux.Unlock()
	sess := &Session{
		m:   make(map[string]any),
		mux: sync.RWMutex{},
		id:  id,
	}
	s.sessions.Set(id, sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	val, exists := s.sessions.Get(id)
	if !exists {
		return ErrSessionNotFound
	}
	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.sessions.Delete(id)
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	sessVal, exists := s.sessions.Get(id)
	sess, ok := sessVal.(session.Session)
	if exists && ok {
		return sess, nil
	}
	return nil, ErrSessionNotFound
}
