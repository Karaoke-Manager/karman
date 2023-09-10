package internal

import (
	"log/slog"
)

type Config struct {
	Debug bool `mapstructure:"debug"`
	Log   struct {
		Level  slog.Level `mapstructure:"level"`
		Format string     `mapstructure:"format"`
	} `mapstructure:"log"`
	DBConnection    string `mapstructure:"db-url"`
	RedisConnection string `mapstructure:"redis-url"`
	API             struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"api"`
	TaskRunner struct {
		Workers int `mapstructure:"workers"`
	} `mapstructure:"task-server"`
	Uploads struct {
		Dir string `mapstructure:"dir"`
	} `mapstructure:"uploads"`
	Media struct {
		Dir string `mapstructure:"dir"`
	} `mapstructure:"media"`
	Cron struct {
		PruneFiles struct {
			Enabled  bool   `mapstructure:"enabled"`
			Schedule string `mapstructure:"schedule"`
		} `mapstructure:"prune-files"`
	} `mapstructure:"cron"`
}
