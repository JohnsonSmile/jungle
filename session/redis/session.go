package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Session struct {
	id     string
	prefix string
	client redis.Cmdable
	script *redis.Script
}

func NewSession(client redis.Cmdable, id string, prefix string) *Session {
	const lua = `
if redis.call("exists", KEY[1])
then
	return redis.call("hset", KEY[1], ARGV[1], ARGV[2])
else
	return -1
end
`
	script := redis.NewScript(lua)
	return &Session{
		client: client,
		id:     id,
		script: script,
		prefix: prefix,
	}
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	key = s.prefix + key
	val, err := s.client.HGet(ctx, s.id, key).Result()
	if err != nil {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, key string, val any) error {
	key = s.prefix + key
	cnt, err := s.script.Run(ctx, s.client, []string{s.id}, key, val).Result()
	if err != nil {
		return err
	}
	if cnt == -1 {
		return ErrKeyNotFound
	}
	return nil
}

func (s *Session) ID() string {
	return s.id
}
