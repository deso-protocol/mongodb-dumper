package cmd

import "github.com/spf13/viper"

type Mode string
type Network string

type Config struct {
	MongoURI        string
	MongoDatabase   string
	MongoCollection string
}

func LoadConfig() *Config {
	config := Config{}

	config.MongoURI = viper.GetString("mongo-uri")
	config.MongoDatabase = viper.GetString("mongo-database")
	config.MongoCollection = viper.GetString("mongo-collection")

	return &config
}
