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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/dakaraj/ptrend/templates"
	_ "github.com/mattn/go-sqlite3" // driver for sqlite3 database
	"github.com/spf13/cobra"
	"github.com/valyala/fasttemplate"
)

var (
	testType     string
	outputPath   string
	testTypeList = []string{"jmeter", "wpt"}
)

// Stats struct contains per-request statistic
type Stats struct {
	Label   string    `json:"label"`
	Average []float64 `json:"average"`
	Max     []float64 `json:"max"`
	Median  []float64 `json:"median"`
	Min     []float64 `json:"min"`
	Perc90  []float64 `json:"perc90"`
	Perc95  []float64 `json:"perc95"`
}

// Results struct represents statistics per-request per-test
type Results struct {
	Tests []string `json:"tests"`
	Stats []Stats  `json:"results"`
}

func convertStatsToFloats(stringStats []string, floatStats []float64) {
	for i, v := range stringStats {
		result, err := strconv.ParseFloat(v, 32)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		floatStats[i] = math.Round(result*100) / 100
	}
}

func generateReport(cmd *cobra.Command, args []string) {
	// defining type of the test
	var testTypeID int
	switch testType {
	case "jmeter":
		testTypeID = 1
	case "wpt":
		testTypeID = 2
	}

	inputPath := args[0]
	var err error
	DB, err = sql.Open("sqlite3", inputPath)
	defer DB.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	rows, err := DB.Query(`
SELECT description
FROM tests
WHERE type_id = ?
ORDER BY test_id ASC;`, testTypeID)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var tests []string
	for rows.Next() {
		var tst string
		rows.Scan(&tst)
		tests = append(tests, tst)
	}

	// ########## JMETER LOGIC ##########
	testsNumber := len(tests)

	rows, err = DB.Query(`
SELECT r.label,
	GROUP_CONCAT(t.description),
	GROUP_CONCAT(r.average),
	GROUP_CONCAT(r.median),
	GROUP_CONCAT(r.perc90),
	GROUP_CONCAT(r.perc95),
	GROUP_CONCAT(r.min),
	GROUP_CONCAT(r.max)
FROM request_statistics AS r
JOIN tests as t ON r.test_id = t.test_id
GROUP BY r.label;
`)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// counting total amount of rows
	row := DB.QueryRow(`SELECT COUNT(DISTINCT label) FROM request_statistics;`)
	var totalRows int
	row.Scan(&totalRows)

	var results Results
	results.Stats = make([]Stats, 0, totalRows)
	for rows.Next() {
		var (
			label          string
			testDecription string
			average        string
			median         string
			perc90         string
			perc95         string
			min            string
			max            string
		)
		rows.Scan(&label, &testDecription, &average, &median, &perc90, &perc95, &min, &max)
		// splitting all concatenated data into arrays
		splitDesc := strings.Split(testDecription, ",")
		splitAverage := strings.Split(average, ",")
		splitMedian := strings.Split(median, ",")
		splitPerc90 := strings.Split(perc90, ",")
		splitPerc95 := strings.Split(perc95, ",")
		splitMin := strings.Split(min, ",")
		splitMax := strings.Split(max, ",")
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
					splitAverage = fillMissingStatValue(i, splitAverage)
					splitMedian = fillMissingStatValue(i, splitMedian)
					splitPerc90 = fillMissingStatValue(i, splitPerc90)
					splitPerc95 = fillMissingStatValue(i, splitPerc95)
					splitMin = fillMissingStatValue(i, splitMin)
					splitMax = fillMissingStatValue(i, splitMax)
				}
			}
		}

		// create struct with calculated metrics
		requestStats := Stats{
			Label:   label,
			Average: make([]float64, testsNumber, testsNumber),
			Median:  make([]float64, testsNumber, testsNumber),
			Perc90:  make([]float64, testsNumber, testsNumber),
			Perc95:  make([]float64, testsNumber, testsNumber),
			Min:     make([]float64, testsNumber, testsNumber),
			Max:     make([]float64, testsNumber, testsNumber),
		}

		// parse each value in array as float and put into an array
		convertStatsToFloats(splitAverage, requestStats.Average)
		convertStatsToFloats(splitMedian, requestStats.Median)
		convertStatsToFloats(splitPerc90, requestStats.Perc90)
		convertStatsToFloats(splitPerc95, requestStats.Perc95)
		convertStatsToFloats(splitMin, requestStats.Min)
		convertStatsToFloats(splitMax, requestStats.Max)

		results.Stats = append(results.Stats, requestStats)
	}

	results.Tests = tests

	// marshall Results struct into JSON
	byteJSON, err := json.Marshal(results)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// take a template string, fill in data and write it as a file
	t := fasttemplate.New(templates.MainJS, "{{", "}}")
	mainJS := t.ExecuteString(map[string]interface{}{
		"data": string(byteJSON),
	})

	// make directory for resuls if needed
	if err := os.MkdirAll(outputPath, os.ModeDir|0755); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// write index.html file
	file, err := os.Create(outputPath + "/index.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	file.WriteString(templates.JmeterTemplate)

	// write main.js file
	file, _ = os.Create(outputPath + "/main.js")
	file.WriteString(mainJS)
}

// validateGenerateArgs function validates arguments for "generate" command
func validateGenerateArgs(cmd *cobra.Command, args []string) error {
	// validate argumets amount
	if len(args) != 1 {
		return errors.New("Please provide path to input DB file as a single argument")
	}

	// validate if input file exists and is not a dir
	if fileInf, err := os.Stat(args[0]); err != nil || fileInf.IsDir() {
		return errors.New("Input file path is invalid or file does not exist")
	}

	// validate if output path is a valid directory
	if outputPath == "" {
		outputPath = "."
	}
	if fileInf, err := os.Stat(outputPath); err == nil && !fileInf.IsDir() {
		return errors.New("Output path is invalid. Should not exist or be a directory")
	}

	// validate source flag is valid
	valid := false
	for _, val := range testTypeList {
		if val == testType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("Test type is not one of the following: %v", testTypeList)
	}

	return nil
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate path/to/db/file",
	Short: "Generate a report from parsed data",
	Long: `Generate command uses data parsed earlier to create a
trends report`,
	Args: validateGenerateArgs,
	Run:  generateReport,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&testType, "source", "s", "jmeter",
		fmt.Sprintf("Chose data source type for report generation: %v", testTypeList))
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Directory to write output file into")
}
