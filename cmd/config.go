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

	config.MongoURI = viper.GetString("monog-uri")
	config.MongoDatabase = viper.GetString("monog-database")
	config.MongoCollection = viper.GetString("monog-collection")

	return &config
}
