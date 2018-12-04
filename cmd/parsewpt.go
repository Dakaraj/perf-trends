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
	"strings"

	"github.com/dakaraj/ptrend/dbutils"
	_ "github.com/mattn/go-sqlite3" // driver for sqlite3 database
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

// WPTResults struct contains all metric data. Tags are used to match
// map keys to struct fields
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

// insertWPTMetrics function executes provided statement with values substitution
func insertWPTMetrics(stmt *sql.Stmt, testID int64, metric string, stats ...interface{}) {
	stats = append([]interface{}{testID, metric}, stats...)
	stmt.Exec(stats...)
}

// parseWPTFiles function parses input file storing results into db file
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

	// decoding an input JSON file into map
	var decodedJSON map[string]interface{}
	decoder := json.NewDecoder(inputFile)
	decoder.Decode(&decodedJSON)
	// leaving only "data" part of the JSON in place
	decodedJSON = decodedJSON["data"].(map[string]interface{})

	wptID := decodedJSON["id"].(string)
	wptLocation := decodedJSON["location"].(string)

	// decoding Average stats
	average := decodedJSON["average"].(map[string]interface{})["firstView"]
	var avgS WPTResults
	mapstructure.Decode(average, &avgS)

	// decoding Standard Deviation stats
	stdDev := decodedJSON["standardDeviation"].(map[string]interface{})["firstView"]
	var stdS WPTResults
	mapstructure.Decode(stdDev, &stdS)

	// decoding Median stats
	median := decodedJSON["median"].(map[string]interface{})["firstView"]
	var medS WPTResults
	mapstructure.Decode(median, &medS)

	description := fmt.Sprintf("%s (%s)", wptID, wptLocation)

	// inserting new test into db getting row id in return
	res, err := DB.Exec(`
INSERT INTO tests (
	description, type_id
) VALUES (
	?, 2
);`, description)
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
INSERT INTO wpt_statistics (
	test_id, metric, responses_200, bytes_out, gzip_savings, requests_full,
	connections, bytes_out_doc, result, base_page_ssl_time, doc_time,
	dom_content_loaded_event_end, image_savings, requests_doc, first_text_paint,
	first_paint, score_cdn, cpu_idle, optimization_checked, image_total,
	score_minify, gzip_total, responses_404, load_time, score_combine,
	first_contentful_paint, first_layout, score_etags
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);`)

	// inserting Average stat metrics
	insertWPTMetrics(insertStatement, lastID, "avg", avgS.Responses200,
		avgS.BytesOut, avgS.GzipSavings, avgS.RequestsFull, avgS.Connections,
		avgS.BytesOutDoc, avgS.Result, avgS.BasePageSSLTime, avgS.DocTime,
		avgS.DomContentLoadedEventEnd, avgS.ImageSavings, avgS.RequestsDoc,
		avgS.FirstTextPaint, avgS.FirstPaint, avgS.ScoreCDN, avgS.CPUIdle,
		avgS.OptimizationChecked, avgS.ImageTotal, avgS.ScoreMinify,
		avgS.GzipTotal, avgS.Responses404, avgS.LoadTime, avgS.ScoreCombine,
		avgS.FirstContentfulPaint, avgS.FirstLayout, avgS.ScoreEtags)

	// inserting Standard Deviation stat metrics
	insertWPTMetrics(insertStatement, lastID, "std", stdS.Responses200,
		stdS.BytesOut, stdS.GzipSavings, stdS.RequestsFull, stdS.Connections,
		stdS.BytesOutDoc, stdS.Result, stdS.BasePageSSLTime, stdS.DocTime,
		stdS.DomContentLoadedEventEnd, stdS.ImageSavings, stdS.RequestsDoc,
		stdS.FirstTextPaint, stdS.FirstPaint, stdS.ScoreCDN, stdS.CPUIdle,
		stdS.OptimizationChecked, stdS.ImageTotal, stdS.ScoreMinify,
		stdS.GzipTotal, stdS.Responses404, stdS.LoadTime, stdS.ScoreCombine,
		stdS.FirstContentfulPaint, stdS.FirstLayout, stdS.ScoreEtags)

	// inserting Median stat metrics
	insertWPTMetrics(insertStatement, lastID, "med", medS.Responses200,
		medS.BytesOut, medS.GzipSavings, medS.RequestsFull, medS.Connections,
		medS.BytesOutDoc, medS.Result, medS.BasePageSSLTime, medS.DocTime,
		medS.DomContentLoadedEventEnd, medS.ImageSavings, medS.RequestsDoc,
		medS.FirstTextPaint, medS.FirstPaint, medS.ScoreCDN, medS.CPUIdle,
		medS.OptimizationChecked, medS.ImageTotal, medS.ScoreMinify,
		medS.GzipTotal, medS.Responses404, medS.LoadTime, medS.ScoreCombine,
		medS.FirstContentfulPaint, medS.FirstLayout, medS.ScoreEtags)
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
}
