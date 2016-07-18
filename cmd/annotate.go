package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/infomodels/datadirectory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// annotateCmd represents the annotate command
var annotateCmd = &cobra.Command{
	Use:   "annotate",
	Short: "annotate data with a metadata descriptor file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		for _, arg := range args {

			var (
				d   *datadirectory.DataDirectory
				err error
			)

			log.WithFields(log.Fields{
				"command":   "annotate",
				"directory": arg,
			}).Info("beginning annotation")

			if d, err = datadirectory.New(&datadirectory.Config{
				DataDirPath:  arg,
				DataVersion:  viper.GetString("dataVersion"),
				Etl:          viper.GetString("etl"),
				Model:        viper.GetString("model"),
				ModelVersion: viper.GetString("modelVersion"),
				Service:      viper.GetString("service"),
				Site:         viper.GetString("site"),
			}); err != nil {
				log.WithFields(log.Fields{
					"command":   "annotate",
					"directory": arg,
					"error":     err,
				}).Fatal("error creating DataDirectory object")
			}

			if err = d.PopulateMetadataFromData(); err != nil {
				log.WithFields(log.Fields{
					"command":   "annotate",
					"directory": arg,
					"error":     err,
				}).Fatal("error populating metadata from data")
			}

			if err = d.WriteMetadataToFile(); err != nil {
				log.WithFields(log.Fields{
					"command":   "annotate",
					"directory": arg,
					"error":     err,
				}).Fatal("error writing metadata to file")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(annotateCmd)

	annotateCmd.Flags().StringP("dataVersion", "d", "", "data version")
	annotateCmd.Flags().StringP("etl", "e", "", "ETL code URL")
	annotateCmd.Flags().StringP("site", "t", "", "site name")
	viper.BindPFlag("dataVersion", annotateCmd.Flags().Lookup("dataVersion"))
	viper.BindPFlag("etl", annotateCmd.Flags().Lookup("etl"))
	viper.BindPFlag("site", annotateCmd.Flags().Lookup("site"))
}
