package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/infomodels/datapackage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// expandCmd represents the expand command
var expandCmd = &cobra.Command{
	Use:   "expand",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		for _, arg := range args {

			var (
				d   *datapackage.DataPackage
				err error
			)

			log.WithFields(log.Fields{
				"command":   "expand",
				"directory": arg,
			}).Info("beginning expand")

			d = &datapackage.DataPackage{
				PackagePath: arg,
				KeyPath:     viper.GetString("keyPath"),
				KeyPassPath: viper.GetString("keyPassPath"),
			}

			if err = d.Unpack(viper.GetString("output")); err != nil {
				log.WithFields(log.Fields{
					"command":   "compress",
					"directory": arg,
					"error":     err,
				}).Fatal("error expanding")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(expandCmd)

	expandCmd.Flags().StringP("keyPath", "k", "", "path to key")
	expandCmd.Flags().StringP("keyPassPath", "p", "", "path to key pass file")
	expandCmd.Flags().StringP("output", "o", "", "output path")
	viper.BindPFlag("keyPath", expandCmd.Flags().Lookup("keyPath"))
	viper.BindPFlag("keyPassPath", expandCmd.Flags().Lookup("keyPassPath"))
	viper.BindPFlag("output", expandCmd.Flags().Lookup("output"))
}
