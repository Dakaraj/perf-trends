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

const testsTable = `
CREATE TABLE IF NOT EXISTS tests (
	test_id INTEGER PRIMARY KEY AUTOINCREMENT,
	description VARCHAR(255) UNIQUE NOT NULL
);`

const requestStatisticsTable = `
CREATE TABLE IF NOT EXISTS request_statistics (
	request_id INTEGER PRIMARY KEY AUTOINCREMENT,
	test_id INT NOT NULL,
	label VARCHAR(255) NOT NULL,
	samples INT NOT NULL,
	average FLOAT NOT NULL,
	median FLOAT NOT NULL,
	perc90 FLOAT NOT NULL,
	perc95 FLOAT NOT NULL,
	min INT NOT NULL,
	max INT NOT NULL,
	FOREIGN KEY (test_id) REFERENCES tests(test_id) ON DELETE CASCADE
);`
