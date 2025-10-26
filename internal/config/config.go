package config

import (
	"time"

	"github.com/hesoyamTM/nbf-auth/internal/adapters/databases/redis"
	"github.com/hesoyamTM/nbf-auth/internal/adapters/oauth2/google"
	"github.com/hesoyamTM/nbf-auth/internal/adapters/oauth2/yandex"
	"github.com/hesoyamTM/nbf-auth/internal/adapters/userclient"
)

type Config struct {
	Env        string                      `yaml:"env" env:"ENV" env-required:"true"`
	Grpc       GRPC                        `yaml:"grpc"`
	App        APP                         `yaml:"app"`
	Redis      redis.RedisConfig           `yaml:"redis"`
	Google     google.GoogleAuthConfig     `yaml:"google"`
	Yandex     yandex.YandexAuthConfig     `yaml:"yandex"`
	UserClient userclient.UserClientConfig `yaml:"user-client"`
}

type APP struct {
	AccessTokenTTL  time.Duration `yaml:"access-token-ttl" env:"ACCESS_TOKEN_TTL" env-required:"true"`
	RefreshTokenTTL time.Duration `yaml:"refresh-token-ttl" env:"REFRESH_TOKEN_TTL" env-required:"true"`
	PrivateKey      string        `env:"PRIVATE_KEY" env-required:"true"`
}

type GRPC struct {
	Host string `yaml:"host" env:"GRPC_HOST" env-required:"true"`
	Port int    `yaml:"port" env:"GRPC_PORT" env-required:"true"`
}
