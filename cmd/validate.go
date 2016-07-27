package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	dms "github.com/chop-dbhi/data-models-service/client"
	validator "github.com/chop-dbhi/data-models-validator"
	"github.com/infomodels/datadirectory"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const sampleSize = 5

var validateCmd = &cobra.Command{
	Use:   "validate [flags] DATADIR",
	Short: "validate a dataset",
	Long: `Validate the checksums and data format of the dataset in DATADIR

Validate the dataset in DATADIR by verifying any metadata passed on the command
line against the metadata file, each checksum in the metadata file against the
appropriate file, and each file against the format prescribed for its table in
the model definition.`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			d   *datadirectory.DataDirectory
			m   *dms.Model
			arg string
			err error
		)

		// Enforce single data directory argument.
		if len(args) != 1 {
			log.WithFields(log.Fields{
				"args": args,
			}).Fatal("validate requires 1 argument")
		}

		arg = args[0]

		log.WithFields(log.Fields{
			"directory": arg,
		}).Info("beginning dataset validation")

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

		// Fill the DataDirectory object with records from the metadata file.
		// Fatal and exit on error.
		if err = d.ReadMetadataFromFile(); err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"error":     err,
			}).Fatal("error reading metadata file")
		}

		// Hack to get Model and ModelVersion actually on to the DataDirectory
		// object, should be done by library.
		// Kevin's hack: command-line args should override file-based configuration
		if d.Model == "" {
			d.Model = d.RecordMaps[0]["cdm"]
		}
		if d.ModelVersion == "" {
			d.ModelVersion = d.RecordMaps[0]["cdm-version"]
		}

		// Validate metadata information and checksums, fatal and exit on err.
		// This checks the attributes on the DataDirectory object (any command
		// line arguments given) against the values in the metadata file and
		// the checksums in the metadata file against the data files.
		if err = d.Validate(); err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"error":     err,
			}).Fatal("error validating checksums and metadata file")
		}

		// Get the model definition for the data format validation.
		if m, err = getModel(d.Model, d.ModelVersion, viper.GetString("service")); err != nil {
			log.WithFields(log.Fields{
				"directory": arg,
				"error":     err,
			}).Fatal("error retrieving data model definition for format validation")
		}

		var (
			hasErrors bool
			table     *dms.Table
		)

		// Iterate over metadata records, running the format validation on each
		// file.
		// TODO: Improve logging in below loop.
		// TODO: Parallelize it!
		// TODO: More comments!
		for _, record := range d.RecordMaps {

			if table = m.Tables.Get(record["table"]); table == nil {
				log.Warnf("* Unknown table '%s'.\nChoices are: %s", record["table"], strings.Join(m.Tables.Names(), ", "))
				continue
			}

			log.Infof("* Evaluating '%s' table in '%s'...", record["table"], record["filename"])

			// Open the reader.
			reader, err := validator.Open(path.Join(d.DirPath, record["filename"]), "")

			if err != nil {
				log.Warnf("* Could not open file: %s", err)
				continue
			}

			v := validator.New(reader, table)

			if err = v.Init(); err != nil {
				log.Warnf("* Problem reading CSV header: %s", err)
				reader.Close()
				continue
			}

			if err = v.Run(); err != nil {
				log.Warnf("* Problem reading CSV data: %s", err)
			}

			reader.Close()

			// Build the result.
			result := v.Result()

			lerrs := result.LineErrors()

			if len(lerrs) > 0 {
				hasErrors = true

				log.Warn("* Row-level issues were found.")

				// Row level issues.
				tw := tablewriter.NewWriter(os.Stdout)

				tw.SetHeader([]string{
					"code",
					"error",
					"occurrences",
					"lines",
					"example",
				})

				var lines, example string

				for err, verrs := range result.LineErrors() {
					ve := verrs[0]

					if ve.Context != nil {
						example = fmt.Sprintf("line %d: `%v` %v", ve.Line, ve.Value, ve.Context)
					} else {
						example = fmt.Sprintf("line %d: `%v`", ve.Line, ve.Value)
					}

					errsteps := errLineSteps(verrs)

					if len(errsteps) > 10 {
						lines = fmt.Sprintf("%s ... (%d more)", strings.Join(errsteps[:10], ", "), len(errsteps[10:]))
					} else {
						lines = strings.Join(errsteps, ", ")
					}

					tw.Append([]string{
						fmt.Sprint(err.Code),
						err.Description,
						fmt.Sprint(len(verrs)),
						lines,
						example,
					})
				}

				tw.Render()
			}

			// Field level issues.
			tw := tablewriter.NewWriter(os.Stdout)

			tw.SetHeader([]string{
				"field",
				"code",
				"error",
				"occurrences",
				"lines",
				"samples",
			})

			var nerrs int

			// Output the error occurrence per field.
			for _, f := range v.Header {
				errmap := result.FieldErrors(f)

				if len(errmap) == 0 {
					continue
				}

				nerrs += len(errmap)

				var (
					lines  string
					sample []*validator.ValidationError
				)

				for err, verrs := range errmap {
					num := len(verrs)

					if num >= sampleSize {
						sample = make([]*validator.ValidationError, sampleSize)

						// Randomly sample.
						for i := range sample {
							j := rand.Intn(num)
							sample[i] = verrs[j]
						}
					} else {
						sample = verrs
					}

					sstrings := make([]string, len(sample))

					for i, ve := range sample {
						if ve.Context != nil {
							sstrings[i] = fmt.Sprintf("line %d: `%s` %s", ve.Line, ve.Value, ve.Context)
						} else {
							sstrings[i] = fmt.Sprintf("line %d: `%s`", ve.Line, ve.Value)
						}
					}

					errsteps := errLineSteps(verrs)

					if len(errsteps) > 10 {
						lines = fmt.Sprintf("%s ... (%d more)", strings.Join(errsteps[:10], ", "), len(errsteps[10:]))
					} else {
						lines = strings.Join(errsteps, ", ")
					}

					tw.Append([]string{
						f,
						fmt.Sprint(err.Code),
						err.Description,
						fmt.Sprint(num),
						lines,
						strings.Join(sstrings, "\n"),
					})
				}
			}

			if nerrs > 0 {
				hasErrors = true
				log.Warn("* Field-level issues were found.")
				tw.Render()
			} else if len(lerrs) == 0 {
				log.Info("* Everything looks good!")
			}

		}

		if hasErrors {
			os.Exit(1)
		}
	},
}

func init() {

	// Register this command under the top-level CLI command.
	RootCmd.AddCommand(validateCmd)

	// Set up the validate-command-specific flags.
	validateCmd.Flags().String("datav", "", "Dataset version number.")
	validateCmd.Flags().String("etl", "", "URL of the ETL code used to create the dataset.")
	validateCmd.Flags().String("site", "", "Name of the organization or site that created the dataset.")

	// Bind viper keys to the flag values.
	viper.BindPFlag("datav", validateCmd.Flags().Lookup("datav"))
	viper.BindPFlag("etl", validateCmd.Flags().Lookup("etl"))
	viper.BindPFlag("site", validateCmd.Flags().Lookup("site"))

}

func getModel(modelName string, versionName string, service string) (*dms.Model, error) {
	// Initialize data models client for service.
	c, err := dms.New(service)

	if err != nil {
		return nil, err
	}

	if err = c.Ping(); err != nil {
		return nil, err
	}

	revisions, err := c.ModelRevisions(modelName)

	if err != nil {
		return nil, err
	}

	var model *dms.Model

	// Get the latest version.
	if versionName == "" {
		model = revisions.Latest()
	} else {

		var (
			versions []string
			_model   *dms.Model
		)

		for _, _model = range revisions.List() {
			if _model.Version == versionName {
				model = _model
				break
			}

			versions = append(versions, _model.Version)
		}

		if model == nil {
			return nil, fmt.Errorf("Invalid version for '%s'. Choose from: %s\n", modelName, strings.Join(versions, ", "))
		}
	}

	log.Infof("Validating against model '%s/%s'", model.Name, model.Version)

	return model, nil
}

// Returns a slice of line ranges that errors have occurred on.
func errLineSteps(errs []*validator.ValidationError) []string {
	var (
		start, end int
		steps      []string
	)

	for _, err := range errs {
		if start == 0 {
			start = err.Line
			end = start
			continue
		}

		if err.Line == end+1 {
			end = err.Line
			continue
		}

		// Skipped a line, log the step
		if start == end {
			steps = append(steps, fmt.Sprint(start))
		} else {
			steps = append(steps, fmt.Sprintf("%d-%d", start, end))
		}

		start = err.Line
		end = err.Line
	}

	// Skipped a line, log the step
	if start == end {
		steps = append(steps, fmt.Sprint(start))
	} else {
		steps = append(steps, fmt.Sprintf("%d-%d", start, end))
	}

	return steps
}
