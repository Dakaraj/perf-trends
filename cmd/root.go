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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// VERSION represents a current version of application
const VERSION = "0.1.2"

var (
	delimiter           string
	header              bool
	ignorePatternString string
	exportFileName      string
	metric              string
	// DB is a packagewide variable that containds database handler
	DB *sql.DB
)

// fillMissingStatValue function inserts a value into array
// at a given index returns modified array
func fillMissingStatValue(index int, stats []string, customVal ...string) []string {
	stats = append(stats, "")
	copy(stats[index+1:], stats[index:])
	if len(customVal) > 0 {
		stats[index] = customVal[0]
	} else {
		stats[index] = "0"
	}

	return stats
}

// getTestsFromDB retrieves tests descriptions from DB for further use
func getTestsFromDB(DB *sql.DB) (tests []string) {
	rows, err := DB.Query(`SELECT description FROM tests ORDER BY test_id ASC;`)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for rows.Next() {
		var tst string
		rows.Scan(&tst)
		tests = append(tests, tst)
	}

	return tests
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ptrend",
	Short: "Parses jmeter log and build performance trends statistics",
	Long: `An application that parses Jmeter log and gathers
by-transaction statistics saving it to SQLite database for further.
Data can be exported as CSV file, or dynamic HTML page.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
