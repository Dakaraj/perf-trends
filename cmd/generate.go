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

	_ "github.com/mattn/go-sqlite3" // driver for sqlite3 database
	"github.com/spf13/cobra"
	"github.com/valyala/fasttemplate"
)

// HTML_PAGE represents a template for a generated page
const HTML_PAGE = `<!DOCTYPE html>
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>Performance Trends Report</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
table, th, td {
    border: 1px solid black;
}
.high {
    background-color: pink;
}
.controls {
    display: inline-block;
}
#comparison-list {
    list-style-type: none;
    margin: 0;
    padding: 0;
    max-height: 100px;
    overflow-y: auto;
    overflow-x: hidden;
    border: 1px solid black;
}
.request-col {
	max-width: 400px;
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
}
#trends-table-body td:nth-child(n+2) {
	width: 120px;
	max-width: 120px;
	text-align: center;
}
#header-row>th:nth-child(n+2) {
	cursor: pointer;
}
    </style>
</head>
<body>
	<div>
        <div class="controls">
            <div>Select metric:</div>
            <select name="metric" id="metric-selector" onchange="rowsPopulate(this.value)">
                <option value="average" selected>Average</option>
                <option value="median">Median</option>
                <option value="perc90">90 Percentile</option>
                <option value="perc95">95 Percentile</option>
                <option value="min">Min</option>
                <option value="max">Max</option>
            </select>
            <div>Compare tests:</div>
            <ul id="comparison-list"></ul>
            <button onclick="compare()">Compare</button>
            <button onclick="resetTable()">Reset</button>
        </div>
		<h1>Average response time per transaction, ms</h1>
        <table id="trends-table">
            <thead>
                <tr id="header-row"></tr>
            </thead>
            <tbody id="trends-table-body"></tbody>
        </table>
    </div>
</body>
<script>
let data = {{data}}

function headerPopulate() {
    let headerRow = document.getElementById("header-row")
    let comparisonList = document.getElementById("comparison-list")
    comparisonList.innerHTML = ""
    headerRow.innerHTML = "<th><span>Test Description</span><hr/><span>Request</span></th>"
    for (let i=0; i<data.tests.length; i++) {
        headerRow.innerHTML += ` + "`<th class=\"desc-header\" onclick=\"sortByColValue(${i})\">${data.tests[i]}</th>`" + `
        comparisonList.innerHTML += ` + "`<li><label for=\"chk${i}\"><input type=\"checkbox\" name=\"chk${i}\" id=\"chk${i}\">${data.tests[i]}</label></li>`" + `
    }
}

function medianCalculator(stats) {
    let rank = 0.5 * (stats.length - 1) + 1
    let ir = Math.floor(rank)
    let fr = rank - ir

	return fr * (stats[ir] - stats[ir-1]) + stats[ir-1]
}

function rowsPopulate(stat) {
	headerPopulate()
    let tBody = document.getElementById("trends-table-body")
    tBody.innerHTML = ""
    let r = data.results
    let rowsMedians = {}
    for (let key in r) {
        let row = ` + "`<tr><td class=\"request-col\" title=\"${key}\">${key}</td>`" + `
        let validValues = r[key][stat].filter((val) => val != 0)
        validValues.sort((a, b) => a - b)
        let rowMedian = medianCalculator(validValues)
        r[key][stat].forEach(function(s) {
			if (s == 0) {
				row += "<td>-</td>"
			} else if (s > rowMedian) {
                row += ` + "`<td class=\"high\">${s}</td>`" + `
            } else {
                row += ` + "`<td>${s}</td>`" + `
            }
        })
        row += "</tr>"
        tBody.innerHTML += row
    }
}
rowsPopulate("average")

function resetTable() {
    headerPopulate()
    rowsPopulate(document.getElementById('metric-selector').value)
}

function compare() {
    let testsArray = document.querySelectorAll("#comparison-list input:checked")
    let metric = document.getElementById("metric-selector").value
    if (testsArray.length < 2) {
        alert("Select at least two tests")
        return
    }
    let idxs = []
    testsArray.forEach(function(elem) {
        idxs.push(parseInt(elem.id.slice(3)))
    })
    let headerRow = document.getElementById("header-row")
    let tBody = document.getElementById("trends-table-body")
    headerRow.innerHTML = "<th><span>Test Description</span><hr/><span>Request</span></th>"
	tBody.innerHTML = ""
	let i = 0
    idxs.forEach(function(idx) {
		headerRow.innerHTML += ` + "`<th class=\"desc-header\" onclick=\"sortByColValue(${i})\">${data.tests[idx]}</th>`" + `
		i++
    })
    let r = data.results
    for (let key in r) {
        let row = ` + "`<tr><td class=\"request-col\" title=\"${key}\">${key}</td>`" + `
        let vals = r[key][metric]
        let baselineVal = undefined
        idxs.forEach(function(idx) {
            let val = vals[idx]
            if (baselineVal === undefined) {
                baselineVal = val
                if (val == 0) {
                    row += "<td>-</td>"
                } else {
                    row += ` + "`<td>${val}</td>`" + `
                }
                return
            }
            if (val == 0) {
                row += "<td>-</td>"
            } else if (baselineVal == 0 || val == baselineVal) {
                row += ` + "`<td>${val}</td>`" + `
            } else {
                let diff = Math.round((val / baselineVal - 1) * 100)
                let color = diff > 0 ? "red" : "green"
                row += ` + "`<td>${val} <span style=\"color: ${color}\">(${diff}%)</span></td>`" + `
            }
        })
        row += "</tr>"
        tBody.innerHTML += row
    }
}

function sortByColValue(idx) {
	let sortClass = document.querySelectorAll("#header-row > th")[idx + 1].classList
	let allRows = document.querySelectorAll("#trends-table-body > tr")
	let allRowsArray = Array.from(allRows)
	let emptyElems = allRowsArray.filter((elem) => elem.childNodes[idx + 1].innerText.trim() == "-")
	allRowsArray = allRowsArray.filter((elem) => elem.childNodes[idx + 1].innerText.trim().match(/([\d\.]+(\s\(-?\d+%\))?)/))
	let orderDesc = sortClass.contains("desc")
	document.querySelectorAll("#header-row > th:nth-child(n+2)").forEach((elem) => elem.classList.remove("desc"))
	let pattern = /^([\d\.]+)/
	allRowsArray.sort(function (a, b) {
		let aText = pattern.exec(a.childNodes[idx + 1].innerText)[1]
		let bText = pattern.exec(b.childNodes[idx + 1].innerText)[1]
		let result = 0
		if (orderDesc) {
			result = aText - bText
		} else {
			result = bText - aText
			sortClass.add("desc")
		}

		return result
	})
	let tableBody = document.querySelector("#trends-table-body")
	tableBody.innerHTML = ""
	allRowsArray.forEach((elem) => tableBody.appendChild(elem))
	emptyElems.forEach((elem) => tableBody.appendChild(elem))
}
</script>
</html>
`

// Results struct represents statistics per-request per-test
type Results struct {
	Tests []string                        `json:"tests"`
	Stats map[string]map[string][]float64 `json:"results"`
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
	inputPath := args[0]
	var err error
	DB, err = sql.Open("sqlite3", inputPath)
	defer DB.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	rows, err := DB.Query(`SELECT description FROM tests ORDER BY test_id ASC;`)
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
	testsNumber := len(tests)

	rows, err = DB.Query(`
SELECT request_statistics.label,
	GROUP_CONCAT(tests.description),
	GROUP_CONCAT(request_statistics.average),
	GROUP_CONCAT(request_statistics.median),
	GROUP_CONCAT(request_statistics.perc90),
	GROUP_CONCAT(request_statistics.perc95),
	GROUP_CONCAT(request_statistics.min),
	GROUP_CONCAT(request_statistics.max)
FROM request_statistics
	JOIN tests ON request_statistics.test_id = tests.test_id
GROUP BY request_statistics.label;
`)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var results Results
	results.Stats = make(map[string]map[string][]float64)
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

		// create a map and arrays for each metric
		requestStats := make(map[string][]float64)
		requestStats["average"] = make([]float64, testsNumber, testsNumber)
		requestStats["median"] = make([]float64, testsNumber, testsNumber)
		requestStats["perc90"] = make([]float64, testsNumber, testsNumber)
		requestStats["perc95"] = make([]float64, testsNumber, testsNumber)
		requestStats["min"] = make([]float64, testsNumber, testsNumber)
		requestStats["max"] = make([]float64, testsNumber, testsNumber)

		// parse each value in array as float and put into an array
		convertStatsToFloats(splitAverage, requestStats["average"])
		convertStatsToFloats(splitMedian, requestStats["median"])
		convertStatsToFloats(splitPerc90, requestStats["perc90"])
		convertStatsToFloats(splitPerc95, requestStats["perc95"])
		convertStatsToFloats(splitMin, requestStats["min"])
		convertStatsToFloats(splitMax, requestStats["max"])

		results.Stats[label] = requestStats
	}

	results.Tests = tests

	// marshall Results struct into JSON
	byteJSON, err := json.Marshal(results)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// take a template string, fill in data and write it as a file
	t := fasttemplate.New(HTML_PAGE, "{{", "}}")
	page := t.ExecuteString(map[string]interface{}{
		"data": string(byteJSON),
	})
	file, err := os.Create("./index.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	file.WriteString(page)
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
}
