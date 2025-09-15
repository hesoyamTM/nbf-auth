package redis

type RedisConfig struct {
	Address  string `yaml:"address" env:"SESSION_REDIS_ADDRESS" env-required:"true"`
	Password string `yaml:"password" env:"SESSION_REDIS_PASSWORD" env-required:"true"`
	DB       int    `yaml:"db" env:"SESSION_REDIS_DB" env-required:"true"`
}
