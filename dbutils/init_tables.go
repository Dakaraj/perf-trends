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

package dbutils

import (
	"database/sql"
)

// Initialize function creates tables in DB based on schemas provided in schemas.go
func Initialize(dbDriver *sql.DB) error {
	statement, err := dbDriver.Prepare(testType)
	if err != nil {
		return err
	}
	statement.Exec()
	dbDriver.Exec(testsTable)
	dbDriver.Exec(requestStatisticsTable)
	dbDriver.Exec(wptStatistics)
	dbDriver.Exec(`INSERT INTO test_types (type_description) VALUES ('load test')`)
	dbDriver.Exec(`INSERT INTO test_types (type_description) VALUES ('web page test')`)

	return nil
}
