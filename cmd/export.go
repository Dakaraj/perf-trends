// Copyright © 2018 Anton Kramarev <kramarev.anton@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // driver for sqlite3 database
	"github.com/spf13/cobra"
)

// metrics variable contains valid values for "metric" flag
var metrics = []string{"average", "median", "perc90", "perc95", "min", "max"}

// exportData function takes data from database and exports it to CSV file
func exportData(cmd *cobra.Command, args []string) {
	inputPath := args[0]
	var err error
	DB, err = sql.Open("sqlite3", inputPath)
	defer DB.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	tests := getTestsFromDB(DB)
	testsNumber := len(tests)

	// depending on "metric" flag value returns a corresponding statistics data
	rows, err := DB.Query(fmt.Sprintf(`
SELECT request_statistics.label,
	GROUP_CONCAT(tests.description),
	GROUP_CONCAT(request_statistics.%s)
FROM request_statistics
	JOIN tests ON request_statistics.test_id = tests.test_id
WHERE tests.type_id = 1
GROUP BY request_statistics.label;
`, metric))
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// create/overwrite an existing file
	file, err := os.Create(exportFileName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fileHandler := csv.NewWriter(file)
	// buffer header to a fileWriter
	fileHandler.Write(append([]string{"Request\\Test"}, tests...))
	for rows.Next() {
		var (
			label        string
			descriptions string
			stats        string
		)
		rows.Scan(&label, &descriptions, &stats)
		// splitting all concatenated data into arrays
		splitDesc := strings.Split(descriptions, ",")
		splitStats := strings.Split(stats, ",")
		// If there is no info for particular transaction in some test
		// then fill all values with zeroes
		if len(splitDesc) != testsNumber {
			diff := testsNumber - len(splitDesc)
			splitDesc = append(splitDesc, make([]string, diff)...)
			for i, v := range tests {
				if v != splitDesc[i] {
					// inserting missing values
					copy(splitDesc[i+1:], splitDesc[i:])
					splitDesc[i] = v
					// insert zero value at current index to each array
					splitStats = fillMissingStatValue(i, splitStats, "")
				}
			}
		}
		// buffer stats line into fileWriter
		fileHandler.Write(append([]string{label}, splitStats...))
	}
	fileHandler.Flush() // write biffered data to a file
}

// validateExportArgs function validates arguments for "export" command
func validateExportArgs(cmd *cobra.Command, args []string) error {
	// validate argumets amount
	if len(args) != 1 {
		return errors.New("Please provide path to input DB file as a single argument")
	}

	// validate if input file exists and is not a dir
	if fileInf, err := os.Stat(args[0]); err != nil || fileInf.IsDir() {
		return errors.New("Input file path is invalid or file does not exist")
	}

	// validate length of delimiter
	if len(delimiter) != 1 {
		return errors.New("Delimiter should only be one character long")
	}

	// validate if output file is not a dir
	if fileInf, err := os.Stat(exportFileName); err == nil && fileInf.IsDir() {
		return errors.New("Output file path is invalid")
	}

	// validate if metric flag has a valid value
	valid := false
	for _, val := range metrics {
		if val == metric {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("Metric is not one of the following: %v", metrics)
	}

	return nil
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export path/to/db/file",
	Short: "Export all trends data to a CSV format",
	Long: `Export all trends data gathered previously to a CSV format.
Can be customized with filename and delimiter.`,
	Args: validateExportArgs,
	Run:  exportData,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "Single character to be used as delimiter")
	exportCmd.Flags().StringVarP(&exportFileName, "name", "n", "export.csv", "Export file name")
	exportCmd.Flags().StringVarP(&metric, "metric", "m", "average", fmt.Sprintf("Select a metric for export: %v", metrics))
}
