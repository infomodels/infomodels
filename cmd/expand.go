package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/infomodels/datapackage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var expandCmd = &cobra.Command{
	Use:   "expand [flags] DATAPACKAGE",
	Short: "expand a data package",
	// TODO: Describe the keyring file better, because IIRC its finicky.
	Long: `Expand DATAPACKAGE using the specified decryption method

Expand the DATAPACKAGE file into the directory specified in the required
output flag. If keypath is given, use the keyring file (ascii encoded private
and public keys) at keypath to decrypt the package. If keypasspath is given,
use its contents as the keyring passcode.`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: Handle errors with more grace.

		var (
			d       *datapackage.DataPackage
			arg     string
			decrypt = false
			err     error
		)

		// Enforce single data package argument.
		if len(args) != 1 {
			log.WithFields(log.Fields{
				"args": args,
			}).Fatal("expand requires 1 argument")
		}

		arg = args[0]

		// Enforce required output path.
		if viper.GetString("output") == "" {
			log.Fatal("expand requires an output path")
		}

		// Determine if decryption will happen for logging.
		if viper.GetString("keypath") != "" {
			decrypt = true
		}

		log.WithFields(log.Fields{
			"package":   arg,
			"decrypt":   decrypt,
			"directory": viper.GetString("output"),
		}).Info("beginning package expansion")

		// On debug, declare that DataPackage object is being created
		// and all the arguements being used.
		log.WithFields(log.Fields{
			"PackagePath": arg,
			"KeyPath":     viper.GetString("keypath"),
			"KeyPassPath": viper.GetString("keypasspath"),
		}).Debug("creating new DataPackage object")

		// Create DataPackage object using all information given. It will
		// handle all the decryption mechanics if those values are given.
		d = &datapackage.DataPackage{
			PackagePath: arg,
			KeyPath:     viper.GetString("keypath"),
			KeyPassPath: viper.GetString("keypasspath"),
		}

		// Fatal and exit if expansion fails.
		if err = d.Unpack(viper.GetString("output")); err != nil {
			log.WithFields(log.Fields{
				"package":   arg,
				"decrypt":   decrypt,
				"directory": viper.GetString("output"),
				"error":     err,
			}).Fatal("error expanding package")
		}

		// Notify that expansion succeeded.
		log.WithFields(log.Fields{
			"directory": viper.GetString("output"),
			"package":   arg,
			"decrypt":   decrypt,
		}).Info("finished package expansion")

	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(expandCmd)

	// Set up the expand-command-specific flags.
	expandCmd.Flags().String("keypath", "", "Path to a keyring file for decryption.")
	expandCmd.Flags().String("keypasspath", "", "Path to a key password file for decryption.")
	expandCmd.Flags().StringP("output", "o", "", "Directory for output. Required.")

	// Bind viper keys to the flag values.
	viper.BindPFlag("keypath", expandCmd.Flags().Lookup("keypath"))
	viper.BindPFlag("keypasspath", expandCmd.Flags().Lookup("keypasspath"))
	viper.BindPFlag("output", expandCmd.Flags().Lookup("output"))
}
