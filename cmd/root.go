package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// This is the top-level cobra command, which represents the binary CLI to the
// user. It defines the most basic "usage" as the name of the binary and adds
// some help text. It will also be used to register the subcommands, in other
// source files.
var RootCmd = &cobra.Command{
	Use:   "infomodels",
	Short: "a healthcare informatics ETL utility",
	Long: `A healthcare informatics ETL utility for chop-dbhi/data-models datasets.

For datasets that conform to a model defined in the chop-dbhi/data-models
repository, this CLI exposes an easy-to-use interface for accomplishing
various ETL-type tasks common to many informatics workflows.`,
}

// Execute initializes, sets up, and runs the CLI. It will run the callbacks
// defined by `cobra.OnInitialize`, set up the options and arguments as defined
// by each command, populate those options and arguments from the command given
// at the command line combined with any environment variable definitions,
// determine which subcommand needs to actually be run, and call Run on it. It
// is called by main.main() in the parent package.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// The log.Fatal automatically calls os.Exit(1) after logging.
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("error initializing infomodels cli")
	}
}

// A Go-native initialization function called after all package imports and
// variable declarations. See https://golang.org/doc/effective_go.html#init
func init() {

	// Add the configuration and log setup callbacks that cobra will call when
	// RootCmd.Execute is called.
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLog)

	// Set up global flags that can be specified at any point on the command
	// line. The values and defaults will be managed by viper.
	RootCmd.PersistentFlags().String("service", "", "Data models service URL.")
	RootCmd.PersistentFlags().String("dmsaservice", "", "Data models SQLAlchemy service URL.")
	RootCmd.PersistentFlags().StringP("model", "m", "", "Data model of the dataset.")
	RootCmd.PersistentFlags().StringP("modelv", "v", "", "Data model version number.")
	RootCmd.PersistentFlags().String("loglvl", "", "Logging output level  [DEBUG|INFO|WARN|ERROR|FATAL].")
	RootCmd.PersistentFlags().String("logfmt", "", "Logging output format [tty|text|json].")

	// Bind viper key names to the global flags.
	viper.BindPFlag("service", RootCmd.PersistentFlags().Lookup("service"))
	viper.BindPFlag("dmsaservice", RootCmd.PersistentFlags().Lookup("dmsaservice"))
	viper.BindPFlag("model", RootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("modelv", RootCmd.PersistentFlags().Lookup("modelv"))
	viper.BindPFlag("loglvl", RootCmd.PersistentFlags().Lookup("loglvl"))
	viper.BindPFlag("logfmt", RootCmd.PersistentFlags().Lookup("logfmt"))

	// Set defaults in viper.
	viper.SetDefault("service", "https://data-models-service.research.chop.edu/")
	viper.SetDefault("dmsaservice", "https://data-models-sqlalchemy.research.chop.edu/")
	viper.SetDefault("loglvl", "INFO")
	if !log.IsTerminal() {
		// Default to json instead of logrus default text when no tty attached.
		viper.SetDefault("logfmt", "json")
	}

	// Set up the dummy version flag. It will actually be handled in the
	// main.main function, but we want it to show up in the help.
	RootCmd.Flags().Bool("version", false, "Show the version and exit.")

}

// initConfig reads in environment variables if set.
func initConfig() {

	// Set up the environment variable configuration. Any environment variable
	// named INFOMODELS_<uppercased viper key name> will be bound to the
	// matching viper key. The environment variable values are overriden by
	// values given at the command line.
	viper.SetEnvPrefix("infomodels")
	viper.AutomaticEnv()

}

// initLog sets up the logging configuration.
func initLog() {

	// Set output to standard error instead of standard out.
	log.SetOutput(os.Stderr)

	// Convert log format argument to proper log formatter. If no log format
	// is given and we are at a tty, logrus will use text and figure out
	// whether colors are supported or not and use them if possible. If not at
	// a tty and no log format is given by the user, we have set it to json in
	// the init function.
	switch viper.GetString("logfmt") {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "tty":
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	case "text":
		log.SetFormatter(&log.TextFormatter{DisableColors: true})
	}

	// Set log level, logging and exiting if an invalid one is given.
	lvl, err := log.ParseLevel(viper.GetString("loglvl"))

	if err != nil {
		fmt.Printf("Error: invalid logging level \"%s\"",
			viper.GetString("loglvl"))
		fmt.Println("Run 'infomodels --help' for usage.")
		log.WithFields(log.Fields{
			"loglvl": viper.GetString("loglvl"),
		}).Fatal("invalid logging level")
	}

	log.SetLevel(lvl)

}
