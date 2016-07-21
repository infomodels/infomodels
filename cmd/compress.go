package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/infomodels/datapackage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var compressCmd = &cobra.Command{
	Use:   "compress [flags] DATADIR",
	Short: "compress data into a package",
	Long: `Compress data in DATADIR into an optionally encrypted package

Compress the dataset in DATADIR into a package at the filename specified with
the output flag. If no output is specified, the data is sent to stdout. If
keyemail is specified, a key registry is searched for a public key and it is
used to encrypt the data. If keypath is specified, the ascii encoded file at
that path is used to encrypt the file. Keypath overrides keyemail and if
neither is given the data is packaged without encryption.`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: Handle errors with more grace.

		var (
			d       *datapackage.DataPackage
			arg     string
			encrypt = false
			err     error
		)

		// Enforce single data directory argument.
		if len(args) != 1 {
			log.WithFields(log.Fields{
				"args": args,
			}).Fatal("compress requires 1 argument")
		}

		arg = args[0]

		// Determine if encryption will happen for logging.
		if viper.GetString("keypath") != "" || viper.GetString("keyemail") != "" {
			encrypt = true
		}

		// Notify that compression is beginning.
		log.WithFields(log.Fields{
			"directory": arg,
			"encrypt":   encrypt,
		}).Info("beginning dataset compression")

		// On debug, declare that DataPackage object is being created
		// and all the arguements being used.
		log.WithFields(log.Fields{
			"PackagePath":    viper.GetString("output"),
			"KeyPath":        viper.GetString("keypath"),
			"PublicKeyEmail": viper.GetString("keyemail"),
		}).Debug("creating new DataPackage object")

		// Create DataPackage object using all information given. It will
		// handle all the encryption mechanics if those values are given.
		d = &datapackage.DataPackage{
			PackagePath:    viper.GetString("output"),
			KeyPath:        viper.GetString("keypath"),
			PublicKeyEmail: viper.GetString("keyemail"),
		}

		// Fatal and exit if compression fails.
		// TODO: Fix .tar.gz compatibility.
		if err = d.Pack(arg); err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"package":   d.PackagePath,
				"encrypt":   encrypt,
				"error":     err,
			}).Fatal("error compressing dataset")
		}

		// Notify that compression succeeded.
		log.WithFields(log.Fields{
			"directory": arg,
			"package":   d.PackagePath,
		}).Info("finished dataset compression")

	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(compressCmd)

	// Set up the compress-command-specific flags.
	compressCmd.Flags().String("keypath", "", "Path to a public key file for encryption.")
	compressCmd.Flags().String("keyemail", "", "Email associated with a public key for encryption.")
	compressCmd.Flags().StringP("output", "o", "", "Compressed package output path.")

	// Bind viper keys to the flag values.
	viper.BindPFlag("keypath", compressCmd.Flags().Lookup("keypath"))
	viper.BindPFlag("keyemail", compressCmd.Flags().Lookup("keyemail"))
	viper.BindPFlag("output", compressCmd.Flags().Lookup("output"))
}
