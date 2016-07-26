package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/infomodels/database"
	"github.com/infomodels/datadirectory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			cfg *datadirectory.Config
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

		// Enforce required schema.
		if viper.GetString("schema") == "" {
			log.Fatal("load requires a schema")
		}

		log.WithFields(log.Fields{
			"directory": arg,
		}).Info("beginning dataset loading")

		// Make the DataDirectory object and load it from the metadata file,
		// see cmd/validate.go for that process.
		cfg = &datadirectory.Config{DataDirPath: arg}
		d, err = datadirectory.New(cfg)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error reading data directory: %v", err))
		}

		err = d.ReadMetadataFromFile()
		if err != nil {
			log.Fatal(fmt.Sprintf("Error reading metadata file: %v", err))
		}

		// The data model in the data directory can be overridden by the
		// --model command line switch. E.g. the non-vocbulary `pedsnet`
		// data model is ordinarily overridden as `--model=pedsnet-core`,
		// and the vocabulary is overridden as `--model=pedsnet-vocab`

		dataModel := d.Model
		if viper.GetString("model") != "" {
			dataModel = viper.GetString("model")
		}

		modelVersion := d.ModelVersion
		if viper.GetString("modelv") != "" {
			modelVersion = viper.GetString("modelv")
		}

		// Open the Database object using information from the DataDirectory
		// object. Create tables and load data, then constraints and indexes.

		// TODO: should we enforce (in validate) that all models and model
		// versions in the metadata file are the same, and in
		// datadirectory, should ReadMetadataFromFile set the model and
		// model version in the datadirectory object? Eh, probably not.
		// We should loop over the metadata, opening the database for each
		// table we load.

		logFields := log.Fields{
			"DataModel":    dataModel,
			"ModelVersion": modelVersion,
			"DbUrl":        viper.GetString("dburi"),
			"Schema":       viper.GetString("schema"),
			"Service":      viper.GetString("service"),
		}

		db, err = database.Open(dataModel, modelVersion, viper.GetString("dburi"), viper.GetString("schema"), viper.GetString("dmsaservice"), "", "")
		if err != nil {
			logFields["err"] = err.Error()
			log.WithFields(logFields).Fatal("Database Open failed")
		}

		err = db.CreateTables()
		if err != nil {
			logFields["err"] = err.Error()
			log.WithFields(logFields).Fatal("CreateTables() failed")
		}

		err = db.Load(d)
		if err != nil {
			logFields["err"] = err.Error()
			log.WithFields(logFields).Fatal("Load() failed")
		}

		// TODO: figure out password handling that does not involve entering
		// it into a command line.
		// TODO: make this command idempotent so that it can wipe any existing
		// data and load the new stuff, to resume from a failure.

	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(loadCmd)

	// Set up the load-command-specific flags.
	loadCmd.Flags().StringP("dburi", "d", "", "Database URI to load the dataset into. Required.")
	loadCmd.Flags().StringP("dbpass", "p", "", "Database password.")
	loadCmd.Flags().StringP("schema", "s", "", "Schema into which to load. Required.")

	// Bind viper keys to the flag values.
	viper.BindPFlag("dburi", loadCmd.Flags().Lookup("dburi"))
	viper.BindPFlag("dbpass", loadCmd.Flags().Lookup("dbpass"))
	viper.BindPFlag("schema", loadCmd.Flags().Lookup("schema"))
}
