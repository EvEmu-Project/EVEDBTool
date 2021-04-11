/*
This function sets up the configuration file
*/

package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("evedb") // name of config file (without extension)
	viper.SetConfigType("yaml")  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")     // optionally look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file missing. Using defaults.")

			//Set Defaults
			viper.SetDefault("log-level", "Info")
			viper.SetDefault("db-host", "127.0.0.1")
			viper.SetDefault("db-port", "3306")
			viper.SetDefault("db-user", "evemu")
			viper.SetDefault("db-pass", "evemu")
			viper.SetDefault("db-database", "evemu")
			viper.SetDefault("migrations-dir", "migrations")
			viper.SetDefault("base-dir", "base")

			//Write configuration file to disk
			viper.WriteConfigAs("./evedb.yaml")
		} else {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}
}
