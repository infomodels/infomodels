package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/infomodels/datapackage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress",
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
				"command":   "compress",
				"directory": arg,
			}).Info("beginning compression")

			d = &datapackage.DataPackage{
				PackagePath:    viper.GetString("output"),
				KeyPath:        viper.GetString("keyPath"),
				PublicKeyEmail: viper.GetString("keyEmail"),
			}

			if err = d.Pack(arg); err != nil {
				log.WithFields(log.Fields{
					"command":   "compress",
					"directory": arg,
					"error":     err,
				}).Fatal("error packing")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(compressCmd)

	compressCmd.Flags().StringP("keyPath", "k", "", "path to public key file")
	compressCmd.Flags().StringP("keyEmail", "e", "", "email of public key")
	compressCmd.Flags().StringP("output", "o", "", "output path")
	viper.BindPFlag("keyPath", compressCmd.Flags().Lookup("keyPath"))
	viper.BindPFlag("keyEmail", compressCmd.Flags().Lookup("keyEmail"))
	viper.BindPFlag("output", compressCmd.Flags().Lookup("output"))
}
