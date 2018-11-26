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
	"os"

	"github.com/dakaraj/perf-trends/dbutils"
	_ "github.com/mattn/go-sqlite3" // driver for sqlite3 database
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

type WPTResults struct {
	Responses200             float64 `mapstructure:"responses_200"`
	BytesOut                 float64 `mapstructure:"bytesOut"`
	GzipSavings              float64 `mapstructure:"gzip_savings"`
	RequestsFull             float64 `mapstructure:"requestsFull"`
	Connections              float64 `mapstructure:"connections"`
	BytesOutDoc              float64 `mapstructure:"bytesOutDoc"`
	Result                   float64 `mapstructure:"result"`
	BasePageSSLTime          float64 `mapstructure:"basePageSSLTime"`
	DocTime                  float64 `mapstructure:"docTime"`
	DomContentLoadedEventEnd float64 `mapstructure:"domContentLoadedEventEnd"`
	ImageSavings             float64 `mapstructure:"image_savings"`
	RequestsDoc              float64 `mapstructure:"requestsDoc"`
	FirstTextPaint           float64 `mapstructure:"firstTextPaint"`
	FirstPaint               float64 `mapstructure:"firstPaint"`
	ScoreCDN                 float64 `mapstructure:"score_cdn"`
	CPUIdle                  float64 `mapstructure:"cpu.Idle"`
	OptimizationChecked      float64 `mapstructure:"optimization_checked"`
	ImageTotal               float64 `mapstructure:"image_total"`
	ScoreMinify              float64 `mapstructure:"score_minify"`
	GzipTotal                float64 `mapstructure:"gzip_total"`
	Responses404             float64 `mapstructure:"responses_404"`
	LoadTime                 float64 `mapstructure:"loadTime"`
	ScoreCombine             float64 `mapstructure:"score_combine"`
	FirstContentfulPaint     float64 `mapstructure:"firstContentfulPaint"`
	FirstLayout              float64 `mapstructure:"firstLayout"`
	ScoreEtags               float64 `mapstructure:"score_etags"`
}

func validateParseWPTArgs(cmd *cobra.Command, args []string) error {
	// validate argumets amount
	if len(args) != 2 {
		return errors.New("Please provide two path arguments")
	}

	// validate if db file is not a dir
	if fileInf, err := os.Stat(args[0]); err == nil && fileInf.IsDir() {
		return errors.New("Output file path is invalid")
	}

	// validate if input file exist and are not a dir
	if fileInf, err := os.Stat(args[1]); err != nil || fileInf.IsDir() {
		return errors.New("Input file path is invalid or file does not exist")
	}

	return nil
}

func parseWPTFiles(cmd *cobra.Command, args []string) {
	outputPath, inputPath := args[0], args[1]
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

	inputFile, err := os.Open(inputPath)
	defer inputFile.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var decodedJSON map[string]interface{}
	decoder := json.NewDecoder(inputFile)
	decoder.Decode(&decodedJSON)
	decodedJSON = decodedJSON["data"].(map[string]interface{})

	testID := decodedJSON["id"].(string)
	testURL := decodedJSON["url"].(string)
	fmt.Println(testID, testURL)

	average := decodedJSON["average"].(map[string]interface{})["firstView"]
	var averageStruct WPTResults
	mapstructure.Decode(average, &averageStruct)

	stdDev := decodedJSON["standardDeviation"].(map[string]interface{})["firstView"]
	var stdDevStruct WPTResults
	mapstructure.Decode(stdDev, &stdDevStruct)

	median := decodedJSON["median"].(map[string]interface{})["firstView"]
	var medianStruct WPTResults
	mapstructure.Decode(median, &medianStruct)

	fmt.Println(averageStruct)
	fmt.Println(stdDevStruct)
	fmt.Println(medianStruct)
}

// parsewptCmd represents the parsewpt command
var parsewptCmd = &cobra.Command{
	Use:   `parsewpt path/to/db/file path/to/input/file`,
	Short: "Parses results JSON file into SQLite database",
	Long: `Parses Web Page Test results file from a provided path
and populates database with new data.`,
	Args: validateParseWPTArgs,
	Run:  parseWPTFiles,
}

func init() {
	rootCmd.AddCommand(parsewptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parsewptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// parsewptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
