package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	Port          int    `yaml:"port"`
	Mode          string `yaml:"mode"`
	EnableSwagger bool   `yaml:"enable_swagger"`
	CorsOrigins   string `yaml:"cors_origins"`
	SwaggerTitle  string `yaml:"swagger_title"`
	SwaggerDesc   string `yaml:"swagger_desc"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type JWTConfig struct {
	Secret        string        `yaml:"secret"`
	Expire        time.Duration `yaml:"expire"`
	RefreshExpire time.Duration `yaml:"refresh_expire"`
}

var C Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &C); err != nil {
		return err
	}
	// 环境变量 SERVER_PORT 优先于 config.yaml，dev.sh 自动换端口时依赖此覆盖
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 && port < 65536 {
			C.Server.Port = port
		}
	}
	return nil
}
