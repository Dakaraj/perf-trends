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
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dakaraj/perf-trends/dbutils"
	_ "github.com/mattn/go-sqlite3" // driver for sqlite3 database
	"github.com/spf13/cobra"
)

var (
	records       = map[string][]int{}
	ignorePattern *regexp.Regexp
)

// RequestStats struct contains statistics data for particular request
type RequestStats struct {
	Label   string
	Samples int
	Average float64
	Median  float64
	Perc90  float64
	Perc95  float64
	Min     int
	Max     int
}

// parseRecord function takes a split line from log, finds label and time elapsed
// then matches label to a provided pattern via "ignore-pattern" flag.
// If label is not matched then parse duration as int and put data into records map
func parseRecord(record []string) {
	label, elapsed := record[2], record[1]
	if ignorePatternString != "" && ignorePattern.MatchString(label) {
		return
	}
	parsedElapsed, _ := strconv.Atoi(elapsed)
	records[label] = append(records[label], parsedElapsed)
}

// calculatePercentile function calculates perentile for values slice provided
func calculatePercentile(stats []int, perc int) float64 {
	if len(stats) == 1 {
		return float64(stats[0])
	}
	rank := float64(perc)/100.0*float64(len(stats)-1) + 1
	ir := int(rank)
	fr := rank - float64(ir)

	percentile := fr*float64(stats[ir]-stats[ir-1]) + float64(stats[ir-1])

	return math.Round(percentile*100) / 100
}

// calculateStats function calculates all metrics and stores them in the struct
func calculateStats(stats []int, rs *RequestStats) {
	length := len(stats)
	sort.Ints(stats)
	var (
		min    = stats[0]
		max    = stats[length-1]
		sum    int
		avg    float64
		median float64
		perc90 float64
		perc95 float64
	)

	for _, v := range stats {
		sum += v
	}

	avg = math.Round(float64(sum)/float64(length)*100) / 100
	median = calculatePercentile(stats, 50)
	perc90 = calculatePercentile(stats, 90)
	perc95 = calculatePercentile(stats, 95)

	rs.Average = avg
	rs.Min = min
	rs.Max = max
	rs.Median = median
	rs.Perc90 = perc90
	rs.Perc95 = perc95
}

// parseJmeterFiles function parses input file storing results into db file
func parseJmeterFiles(cmd *cobra.Command, args []string) {
	ignorePattern = regexp.MustCompile(ignorePatternString)
	description, outputPath, inputPaths := args[0], args[1], args[2:]
	// removing all commas as those are used for concatenation later
	description = strings.Replace(description, ",", "", -1)
	var err error
	DB, err = sql.Open("sqlite3", outputPath)
	defer DB.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := dbutils.Initialize(DB); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, inputPath := range inputPaths {
		inputFile, err := os.Open(inputPath)
		defer inputFile.Close()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		reader := csv.NewReader(inputFile)
		if delimiter != "," {
			reader.Comma = rune(delimiter[0])
		}

		// skipping first line of log as header
		if header {
			reader.Read()
		}

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			parseRecord(record)
		}
	}

	// inserting new test into db getting row id in return
	res, err := DB.Exec(`
INSERT INTO tests
	(description, type_id)
	VALUES (?, 1);
`, description)
	if err != nil {
		// stop process if description is not unique
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			fmt.Println("Provided test description is not unique")
		} else {
			fmt.Println(err.Error())
		}
		os.Exit(1)
	}
	lastID, _ := res.LastInsertId()

	// preparing an insert statement
	insertStatement, _ := DB.Prepare(`
INSERT INTO request_statistics
	(test_id, label, samples, average, median, perc90, perc95, min, max)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`)
	for req, stats := range records {
		rs := RequestStats{}
		rs.Label = req
		rs.Samples = len(stats)
		calculateStats(stats, &rs)
		// insert data in db row by row
		_, err := insertStatement.Exec(lastID, rs.Label, rs.Samples, rs.Average,
			rs.Median, rs.Perc90, rs.Perc95, rs.Min, rs.Max)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

// validateParseArgs function validates arguments for "parsejmeter" command
func validateParseJmeterArgs(cmd *cobra.Command, args []string) error {
	// validate argumets amount
	if len(args) < 3 {
		return errors.New(
			"Please provide a unique description and at least two path arguments",
		)
	}

	// validate if db file is not a dir
	if fileInf, err := os.Stat(args[1]); err == nil && fileInf.IsDir() {
		return errors.New("Output file path is invalid")
	}

	// validate if input files exist and are not a dir
	for _, val := range args[2:] {
		if fileInf, err := os.Stat(val); err != nil || fileInf.IsDir() {
			return errors.New("Input file path is invalid or file does not exist")
		}
	}

	// validate length of delimiter
	if len(delimiter) != 1 {
		return errors.New("Delimiter should only be one character long")
	}

	// validate provided ignore pattern
	if _, err := regexp.Compile(ignorePatternString); err != nil {
		return errors.New("Provided ignore pattern is invalid")
	}

	return nil
}

// parsejmeterCmd represents the parsejmeter command
var parsejmeterCmd = &cobra.Command{
	Use:   `parsejmeter "unique test description" path/to/db/file path/to/input/file [other/input/files...]`,
	Short: "Parses Jmeter log file into SQLite database",
	Long: `Parses Jmeter log file from a provided path and populates
database with new data.`,
	Args: validateParseJmeterArgs,
	Run:  parseJmeterFiles,
}

func init() {
	rootCmd.AddCommand(parsejmeterCmd)

	parsejmeterCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "Single character to be used as delimiter")
	parsejmeterCmd.Flags().BoolVarP(&header, "field-names", "f", false, "Use if input file contains a header line with field names")
	parsejmeterCmd.Flags().StringVarP(&ignorePatternString, "ignore-pattern", "i", "", "Label regex pattern that will be ignored by parser")
}
