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

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			d   *datadirectory.DataDirectory
			m   *dms.Model
			err error
		)

		for _, arg := range args {

			log.WithFields(log.Fields{
				"command":   "validate",
				"directory": arg,
			}).Info("beginning validation")

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
					"command":   "validate",
					"directory": arg,
					"error":     err,
				}).Fatal("error creating DataDirectory object")
			}

			if err = d.Validate(); err != nil {
				log.WithFields(log.Fields{
					"command":   "validate",
					"directory": arg,
					"error":     err,
				}).Fatal("error validating metadata file")
			}

			// Actually fill the DataDirectory object with records.
			if err = d.ReadMetadataFromFile(); err != nil {
				log.Fatal(err)
			}

			// Hack to get Model and ModelVersion actually on to the DataDirectory object, should be done by library.
			d.Model = d.RecordMaps[0]["cdm"]
			d.ModelVersion = d.RecordMaps[0]["cdm-version"]

			if m, err = getModel(d.Model, d.ModelVersion, viper.GetString("service")); err != nil {
				log.Fatal(err)
			}

			var (
				hasErrors bool
				table     *dms.Table
			)

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
							for i, _ := range sample {
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
		}
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringP("dataVersion", "d", "", "data version")
	validateCmd.Flags().StringP("etl", "e", "", "ETL code URL")
	validateCmd.Flags().StringP("site", "t", "", "site name")
	viper.BindPFlag("dataVersion", validateCmd.Flags().Lookup("dataVersion"))
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
