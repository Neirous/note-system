package config

type Config struct {
	Server ServerConfig `yaml:"server"`
	Mysql  MysqlConfig  `yaml:"mysql"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type MysqlConfig struct {
	Dsn string `yaml:"dsn"`
}
