package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile           string
	verbose           bool
	logJSON           bool
	defaultServiceURL = "https://data-models-service.research.chop.edu"
)

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "infomodels",
	Short: "informatics data model utility",
	Long: `Infomodels is a healthcare informatics data model utility.

It includes functionality for packaging and validating data sets.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLog)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.infomodels.yaml)")
	RootCmd.PersistentFlags().StringP("service", "s", "https://data-models-service.research.chop.edu", "data models service URL")
	RootCmd.PersistentFlags().StringP("model", "m", "", "data model name")
	RootCmd.PersistentFlags().StringP("modelVersion", "n", "", "data model version")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolVarP(&logJSON, "logJSON", "j", false, "log JSON")
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("service", RootCmd.PersistentFlags().Lookup("service"))
	viper.BindPFlag("model", RootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("modelVersion", RootCmd.PersistentFlags().Lookup("modelVersion"))
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("logJSON", RootCmd.PersistentFlags().Lookup("logJSON"))
	viper.SetDefault("service", defaultServiceURL)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".infomodels") // name of config file (without extension)
	viper.AddConfigPath(".")           // look for config in working directory first
	viper.AddConfigPath("$HOME")       // adding home directory as search path
	viper.SetEnvPrefix("infomodels")   // add prefix to environment variable names
	viper.AutomaticEnv()               // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			log.Fatalf("error using config file '%s': %s \n", viper.ConfigFileUsed(), err)
		}
	} else {
		log.Printf("using config file: '%s'", viper.ConfigFileUsed())
	}
}

// initLog sets up the logging configuration.
func initLog() {
	log.SetLevel(log.WarnLevel)
	log.SetOutput(os.Stderr)

	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	if viper.GetBool("logJSON") {
		log.SetFormatter(&log.JSONFormatter{})
	}
}
