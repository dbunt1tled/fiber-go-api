package config

import (
	"log/slog"
	"time"
)

type Config struct {
	Name      string       `koanf:"name"`
	URL       string       `koanf:"url"`
	Env       string       `koanf:"env"`
	Debug     bool         `koanf:"debug"`
	Profiling bool         `koanf:"profiling"`
	Server    ServerConfig `koanf:"server"`
	DB        DBConfig     `koanf:"db"`
	Redis     RedisConfig  `koanf:"redis"`
	Log       LogConfig    `koanf:"log"`
	Mailer    MailerConfig `koanf:"mailer"`
	Static    StaticConfig `koanf:"static"`
}
type ServerConfig struct {
	HTTP HTTPConfig `koanf:"http"`
	JWT  JWTConfig  `koanf:"jwt"`
}

type HTTPConfig struct {
	Host      string        `koanf:"host"`
	Port      int           `koanf:"port"`
	Timeout   time.Duration `koanf:"timeout"`
	Prefork   bool          `koanf:"prefork"`
	CORS      CORSConfig    `koanf:"cors"`
	BodyLimit int           `koanf:"bodylimit"`
	TLS       TLSConfig     `koanf:"tls"`
}

type TLSConfig struct {
	Keyfile  string `koanf:"keyfile"`
	CertFile string `koanf:"certfile"`
}

type JWTConfig struct {
	PublicKey  string     `koanf:"public"`
	PrivateKey string     `koanf:"private"`
	Algorithm  string     `koanf:"algorithm"`
	Expire     ExpireConf `koanf:"expire"`
}

type ExpireConf struct {
	Access  time.Duration `koanf:"access"`
	Refresh time.Duration `koanf:"refresh"`
	Confirm time.Duration `koanf:"confirm"`
}

type LogConfig struct {
	Level slog.Level `koanf:"level"`
	File  string     `koanf:"file"`
}
type CORSConfig struct {
	AllowMethods  string `koanf:"allowmethods"`
	AllowHeaders  string `koanf:"allowheaders"`
	AllowOrigins  string `koanf:"alloworigins"`
	ExposeHeaders string `koanf:"exposeheaders"`
}

type DBConfig struct {
	Main MainDBConfig `koanf:"main"`
}

type RedisConfig struct {
	Addr string `koanf:"addr"`
}

type MainDBConfig struct {
	DSN string `koanf:"dsn"`
}

type MailerConfig struct {
	Address  string `koanf:"address"`
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type StaticConfig struct {
	URL       string `koanf:"url"`
	Directory string `koanf:"dir"`
}
