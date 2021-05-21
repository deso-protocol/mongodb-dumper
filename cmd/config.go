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

	config.MongoURI = viper.GetString("mongodb-uri")
	config.MongoDatabase = viper.GetString("mongodb-database")
	config.MongoCollection = viper.GetString("mongodb-collection")

	return &config
}
