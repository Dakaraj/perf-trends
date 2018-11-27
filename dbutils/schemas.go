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

const testType = `
CREATE TABLE IF NOT EXISTS test_types (
	type_id INTEGER PRIMARY KEY AUTOINCREMENT,
	type_description VARCHAR(64)
);`

const testsTable = `
CREATE TABLE IF NOT EXISTS tests (
	test_id INTEGER PRIMARY KEY AUTOINCREMENT,
	description VARCHAR(255) UNIQUE NOT NULL,
	type_id INT NOT NULL,
	FOREIGN KEY (type_id) REFERENCES test_types(type_id) ON DELETE CASCADE
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

const wptStatistics = `
CREATE TABLE IF NOT EXISTS wpt_statistics (
	wpt_id INTEGER PRIMARY KEY AUTOINCREMENT,
	test_id INT NOT NULL,
	metric VARCHAR(3) CHECK (metric IN ('avg', 'std', 'med')) NOT NULL,
	responses_200 FLOAT NOT NULL,
	bytes_out FLOAT NOT NULL,
	gzip_savings FLOAT NOT NULL,
	requests_full FLOAT NOT NULL,
	connections FLOAT NOT NULL,
	bytes_out_doc FLOAT NOT NULL,
	result FLOAT NOT NULL,
	base_page_ssl_time FLOAT NOT NULL,
	doc_time FLOAT NOT NULL,
	dom_content_loaded_event_end FLOAT NOT NULL,
	image_savings FLOAT NOT NULL,
	requests_doc FLOAT NOT NULL,
	first_text_paint FLOAT NOT NULL,
	first_paint FLOAT NOT NULL,
	score_cdn FLOAT NOT NULL,
	cpu_idle FLOAT NOT NULL,
	optimization_checked FLOAT NOT NULL,
	image_total FLOAT NOT NULL,
	score_minify FLOAT NOT NULL,
	gzip_total FLOAT NOT NULL,
	responses_404 FLOAT NOT NULL,
	load_time FLOAT NOT NULL,
	score_combine FLOAT NOT NULL,
	first_contentful_paint FLOAT NOT NULL,
	first_layout FLOAT NOT NULL,
	score_etags FLOAT NOT NULL,
	FOREIGN KEY (test_id) REFERENCES tests(test_id) ON DELETE CASCADE
);`
