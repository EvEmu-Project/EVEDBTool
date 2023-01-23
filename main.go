/*
This is a simple program for the recording of Chaturbate live streams.
*/

package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log = logrus.New()

func main() {
	os.Exit(realMain())
}

var ui cli.Ui

var version = "dev"

func realMain() int {
	//Define version and print on startup
	log.Info("EVEDBTool ", version)

	//Initialize config, logging and closehandler
	initConfig()
	setupLogging()

	ui = &cli.BasicUi{Writer: os.Stdout}

	cli := &cli.CLI{
		Args: os.Args[1:],
		Commands: map[string]cli.CommandFactory{
			"install": func() (cli.Command, error) {
				return &InstallCommand{}, nil
			},
			"up": func() (cli.Command, error) {
				return &UpCommand{}, nil
			},
			"down": func() (cli.Command, error) {
				return &DownCommand{}, nil
			},
			"redo": func() (cli.Command, error) {
				return &RedoCommand{}, nil
			},
			"status": func() (cli.Command, error) {
				return &StatusCommand{}, nil
			},
			"new": func() (cli.Command, error) {
				return &NewCommand{}, nil
			},
			"skip": func() (cli.Command, error) {
				return &SkipCommand{}, nil
			},
			"seed": func() (cli.Command, error) {
				return &SeedCommand{}, nil
			},
			"dungeon": func() (cli.Command, error) {
				return &DungeonCommand{}, nil
			},
			"dungeon list": func() (cli.Command, error) {
				return &DungeonListCommand{}, nil
			},
			"dungeon apply": func() (cli.Command, error) {
				return &DungeonApplyCommand{}, nil
			},
			"dungeon import": func() (cli.Command, error) {
				return &DungeonImportCommand{}, nil
			},
			"dungeon export": func() (cli.Command, error) {
				return &DungeonExportCommand{}, nil
			},
			"dungeon new": func() (cli.Command, error) {
				return &DungeonNewCommand{}, nil
			},
			"dungeon delete": func() (cli.Command, error) {
				return &DungeonDeleteCommand{}, nil
			},
			"dungeon add-room": func() (cli.Command, error) {
				return &DungeonAddRoomCommand{}, nil
			},
			"dungeon remove-room": func() (cli.Command, error) {
				return &DungeonRemoveRoomCommand{}, nil
			},
			"dungeon list-factions": func() (cli.Command, error) {
				return &DungeonFactionListCommand{}, nil
			},
			"dungeon list-archetypes": func() (cli.Command, error) {
				return &DungeonArchetypeListCommand{}, nil
			},
			"dungeon list-rooms": func() (cli.Command, error) {
				return &DungeonRoomListCommand{}, nil
			},
		},
		HelpFunc: cli.BasicHelpFunc("evedbtool"),
		Version:  version,
	}

	exitCode, err := cli.Run()
	if err != nil {
		log.Error(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}

//Sets up logging levels based upon configuration file
func setupLogging() {
	log_level := viper.Get("log-level")

	if log_level == "Debug" {
		log.Level = logrus.DebugLevel
	} else if log_level == "Trace" {
		log.Level = logrus.TraceLevel
	} else if log_level == "Warn" {
		log.Level = logrus.WarnLevel
	} else if log_level == "Error" {
		log.Level = logrus.ErrorLevel
	} else {
		log.Level = logrus.InfoLevel
	}
}
