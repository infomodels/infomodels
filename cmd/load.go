package cmd

import (
	"github.com/infomodels/database"
	"github.com/infomodels/datadirectory"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load [flags] DATADIR",
	Short: "load a dataset into a database",
	Long: `Load the dataset in DATADIR into the dburi specified database

Load the dataset in DATADIR by using the metadata to determine which file to
load into which table. The model tables are created, data loaded, and then the
constraints and finally the indexes are created.`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			d   *datadirectory.DataDirectory
			db  *database.Database
			arg string
			err error
		)

		// Enforce single data directory argument.
		if len(args) != 1 {
			log.WithFields(log.Fields{
				"args": args,
			}).Fatal("load requires 1 argument")
		}

		arg = args[0]

		// Enforce required dburi.
		if viper.GetString("dburi") == "" {
			log.Fatal("load requires a dburi")
		}

		log.WithFields(log.Fields{
			"directory": arg,
		}).Info("beginning dataset loading")

		// Make the DataDirectory object and load it from the metadata file,
		// see cmd/validate.go for that process.

		// Open the Database object using information from the DataDirectory
		// object. Create tables and load data, then constraints and indexes.

		// TODO: figure out password handling that does not involve entering
		// it into a command line.
		// TODO: make this command idempotent so that it can wipe any existing
		// data and load the new stuff, to resume from a failure.

	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(loadCmd)

	// Set up the expand-command-specific flags.
	loadCmd.Flags().StringP("dburi", "d", "", "Database URI to load the dataset into. Required.")
	loadCmd.Flags().StringP("dbpass", "p", "", "Database password.")

	// Bind viper keys to the flag values.
	viper.BindPFlag("dburi", loadCmd.Flags().Lookup("dburi"))
	viper.BindPFlag("dbpass", loadCmd.Flags().Lookup("dbpass"))
}
