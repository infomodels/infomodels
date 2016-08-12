package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/infomodels/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

var constrainCmd = &cobra.Command{
	Use:   "constrain [flags]",
	Short: "add constraints to a data model instance",
	Long: `Add foreign key constraints to a data model instance.

Add foreign key constraints to the data model instance located in the
database specified by the dburi switch and further specified by the
searchPath switch. The model and model version are looked up in the
database in the version_history table.

The required searchPath switch is a PostgreSQL search_path value
containing a comma-separated list of schema names. The first schema in
the list is the primary schema in which the constraints will be
added. Additional schemas may be required if the tables have
foreign keys into tables located in other schemas.`,

	Run: func(cmd *cobra.Command, args []string) {

		var (
			db           *database.Database
			dburi        string
			searchPath   string
			dmsaservice  string
			dataModel    string
			modelVersion string
			err          error
		)

		// Enforce required dburi.
		dburi = viper.GetString("dburi")
		if dburi == "" {
			log.Fatal("constrain requires a dburi")
		}

		// Enforce required searchPath.
		searchPath = viper.GetString("searchPath")
		if searchPath == "" {
			log.Fatal("constrain requires a searchPath")
		}

		dmsaservice = viper.GetString("dmsaservice")

		dataModel, modelVersion, err = getModelAndVersion(dburi, searchPath)
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error()}).Fatal("Failed to get model and version")
		}

		if viper.GetString("model") != "" {
			dataModel = viper.GetString("model")
		}

		if viper.GetString("modelv") != "" {
			modelVersion = viper.GetString("modelv")
		}

		logFields := log.Fields{
			"dataModel":    dataModel,
			"modelVersion": modelVersion,
			"dburi":        dburi,
			"searchPath":   searchPath,
			"dmsaservice":  dmsaservice,
		}

		db, err = database.Open(dataModel, modelVersion, dburi, searchPath, dmsaservice, "", "")
		if err != nil {
			logFields["err"] = err.Error()
			log.WithFields(logFields).Fatal("Database Open failed")
		}

		// TODO: add switches for database error sensitivities (normal/strict/force)

		if !viper.GetBool("undo") {

			log.WithFields(logFields).Info("adding foreign key constraints")

			constraintsStart := time.Now()
			err = db.CreateConstraints("normal")
			if err != nil {
				elapsed := time.Since(constraintsStart)
				logFields["durationMinutes"] = elapsed.Minutes()
				logFields["err"] = err.Error()
				log.WithFields(logFields).Fatal("error while adding constraints")
			}

			elapsed := time.Since(constraintsStart)
			logFields["durationMinutes"] = elapsed.Minutes()
			log.WithFields(logFields).Info("constraints added")

		} else {

			// Drop constraints
			log.WithFields(logFields).Info("dropping constraints")

			constraintsStart := time.Now()
			err = db.DropConstraints("normal")
			if err != nil {
				elapsed := time.Since(constraintsStart)
				logFields["durationMinutes"] = elapsed.Minutes()
				logFields["err"] = err.Error()
				log.WithFields(logFields).Fatal("error while dropping constraints")
			}

			elapsed := time.Since(constraintsStart)
			logFields["durationMinutes"] = elapsed.Minutes()
			log.WithFields(logFields).Info("constraints dropped")

		}

	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(constrainCmd)

	// Made these into global flags because I couldn't figure out how to use them in another subcommand also.
	// // Set up the constrain-command-specific flags.
	// constrainCmd.Flags().StringP("dburi", "d", "", "Database URI. Required.")
	// //	constrainCmd.Flags().StringP("dbpass", "p", "", "Database password.")
	// constrainCmd.Flags().StringP("searchPath", "s", "", "SearchPath indicating primary and optional secondary schemas. Required.")
	// constrainCmd.Flags().Bool("undo", false, "Undo/drop all constraints.")

	// // Bind viper keys to the flag values.
	// viper.BindPFlag("dburi", constrainCmd.Flags().Lookup("dburi"))
	// //	viper.BindPFlag("dbpass", constrainCmd.Flags().Lookup("dbpass"))
	// viper.BindPFlag("searchPath", constrainCmd.Flags().Lookup("searchPath"))
	// viper.BindPFlag("undo", constrainCmd.Flags().Lookup("undo"))
}
