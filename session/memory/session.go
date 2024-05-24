package memory

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Session struct {
	m   map[string]any
	mux sync.RWMutex
	id  string
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	val, exists := s.m[key]
	if !exists {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, key string, val any) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.m[key] = val
	return nil
}

func (s *Session) ID() string {
	return s.id
}
