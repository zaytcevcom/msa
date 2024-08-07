package main

import "github.com/spf13/viper"

type Config struct {
	Logger   loggerConf
	Postgres postgresConf
	Rabbit   rabbitConf
	Redis    redisConf
}

type loggerConf struct {
	Level string
}

type postgresConf struct {
	Dsn string
}

type rabbitConf struct {
	URI      string
	Exchange string
	Queue    string
}

type redisConf struct {
	Host     string
	Port     int
	Password string
}

func LoadConfig(path string) (Config, error) {
	config := Config{}

	viper.SetConfigFile(path)

	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return config, err
}
