package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/infomodels/datadirectory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var annotateCmd = &cobra.Command{
	Use:   "annotate [flags] DATADIR",
	Short: "annotate data with a metadata file",
	Long: `Collect metadata for and write a metadata file to DATADIR

Collect metadata by prompting the user to input any values not given in the
flags and then calculating the checksums of data files in the directory. Data
files are matched to tables in the data model by name and the user is prompted
to enter the table for any data file with a name that does not match a table
in the model. Any existing metadata file in the directory is overwritten.`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: Handle errors with more grace.

		var (
			d   *datadirectory.DataDirectory
			arg string
			err error
		)

		// Enforce single data directory argument.
		if len(args) != 1 {
			log.WithFields(log.Fields{
				"args": args,
			}).Fatal("annotate requires 1 argument")
		}

		arg = args[0]

		// Notify that dataset annotation is beggining on this directory.
		log.WithFields(log.Fields{
			"directory": arg,
		}).Info("beginning dataset annotation")

		// On debug, declare that DataDirectory object is being created
		// and all the arguements being used.
		log.WithFields(log.Fields{
			"DataDirPath":  arg,
			"DataVersion":  viper.GetString("datav"),
			"Etl":          viper.GetString("etl"),
			"Model":        viper.GetString("model"),
			"ModelVersion": viper.GetString("modelv"),
			"Service":      viper.GetString("service"),
			"Site":         viper.GetString("site"),
		}).Debug("creating new DataDirectory object")

		// Create DataDirectory object, using all information given.
		d, err = datadirectory.New(&datadirectory.Config{
			DataDirPath:  arg,
			DataVersion:  viper.GetString("datav"),
			Etl:          viper.GetString("etl"),
			Model:        viper.GetString("model"),
			ModelVersion: viper.GetString("modelv"),
			Service:      viper.GetString("service"),
			Site:         viper.GetString("site"),
		})

		// Fatal and exit if the object creation fails.
		if err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"error":     err,
			}).Fatal("error creating DataDirectory object")
		}

		// On debug, notify that metadata population is about to occur.
		log.WithFields(log.Fields{
			"directory": arg,
		}).Debug("populating metadata from data and user input")

		// This function asks for user input, checking it against the data
		// model, to fill in the needed metadata values and then iterates
		// over the files in the directory. If the file name doesn't match
		// a table in the model, the user is asked to input the table name.
		// The checksums are calculated for each file.
		// TODO: Move the metadata filling routines up to this level.
		if err = d.PopulateMetadataFromData(); err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"error":     err,
			}).Fatal("error populating metadata")
		}

		// On debug, notify that the file is being written.
		log.WithFields(log.Fields{
			"directory": arg,
			"file":      d.FilePath,
		}).Debug("writing metadata to file")

		// This function simply writes the csv metadata file into the data
		// directory. It overwrites any existing file.
		if err = d.WriteMetadataToFile(); err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"error":     err,
			}).Fatal("error writing metadata to file")
		}

		// Notify that the process is complete for this directory.
		log.WithFields(log.Fields{
			"directory": arg,
		}).Info("finished dataset annotation")

	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(annotateCmd)

	// Set up the annotate-command-specific flags.
	annotateCmd.Flags().String("datav", "", "Dataset version number.")
	annotateCmd.Flags().String("etl", "", "URL of the ETL code used to create the dataset.")
	annotateCmd.Flags().String("site", "", "Name of the organization or site that created the dataset.")

	// Bind viper keys to the flag values.
	viper.BindPFlag("datav", annotateCmd.Flags().Lookup("datav"))
	viper.BindPFlag("etl", annotateCmd.Flags().Lookup("etl"))
	viper.BindPFlag("site", annotateCmd.Flags().Lookup("site"))

}
