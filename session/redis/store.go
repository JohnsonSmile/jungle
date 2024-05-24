package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"jungle/session"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not exists")
)

type Store struct {
	client     redis.Cmdable
	expiration time.Duration
	prefix     string
}

func NewStore(client redis.Cmdable, expiration time.Duration, prefix string) *Store {
	return &Store{
		client:     client,
		expiration: expiration,
		prefix:     prefix,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	_, err := s.client.HSet(ctx, id, id, id).Result()
	if err != nil {
		return nil, err
	}
	_, err = s.client.Expire(ctx, id, s.expiration).Result()
	if err != nil {
		return nil, err
	}
	return NewSession(s.client, id, s.prefix), nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	ok, err := s.client.Expire(ctx, id, s.expiration).Result()
	if !ok {
		return ErrSessionNotFound
	}
	return err
}

func (s *Store) Remove(ctx context.Context, id string) error {
	cnt, err := s.client.Del(ctx, id).Result()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	cnt, err := s.client.Exists(ctx, id).Result()
	if err != nil || cnt != 1 {
		return nil, ErrSessionNotFound
	}
	return NewSession(s.client, id, s.prefix), nil
}
