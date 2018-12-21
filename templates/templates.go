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

package templates

// JmeterTemplate represents a template for a generated page with Jmeter stats
const JmeterTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
    <title>Performance Trends Report</title>
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <script src="https://d3js.org/d3.v5.min.js"></script>
    <style>
      table,
      th,
      td {
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
      tr > td:first-child,
      tr > th:first-child {
        width: 400px;
        max-width: 400px;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
      }
      #trends-table-body td:nth-child(n + 2) {
        width: 120px;
        max-width: 120px;
        text-align: center;
      }
      #header-row > th:nth-child(n + 2) {
        cursor: pointer;
      }
      .bar-container {
        padding: 2px 0px 4px 0px;
        position: absolute;
        display: block;
        height: 100px;
        border: 2px solid gray;
        min-width: 20px;
        background-color: beige;
      }
      .bar {
        display: inline-block;
        background-color: green;
        width: 15px;
        border: 1px solid darkgreen;
        margin: 0 2px;
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
      <div class="bar-container" style="display: none"></div>
      <h1>Average response time per transaction, ms</h1>
      <table id="trends-table">
        <thead>
          <tr id="header-row"></tr>
        </thead>
        <tbody id="trends-table-body"></tbody>
      </table>
    </div>
  </body>
  <script src="main.js"></script>
</html>
`

// MainJS contains page main logic
const MainJS = `let data = {{data}};

function headerPopulate() {
    let headerRow = d3.select("#header-row");
    let comparisonList = d3.select("#comparison-list");

    comparisonList.html("");
    headerRow.html(
        "<th><span>Test Description</span><hr/><span>Request</span></th>"
    );

    headerRow
        .selectAll("th:nth-child(n+2)")
        .data(data.tests)
        .enter()
    .append("th")
        .on("click", function(_, i) {
            sortByColValue(i);
        })
        .text(d => d);

    comparisonList
        .selectAll("li")
        .data(data.tests)
        .enter()
    .append("li")
        .html(
            (d, i) =>
                ` + "`<label for=\"chk${i}\"><input type=\"checkbox\" name=\"chk${i}\" id=\"chk${i}\">${d}</label>`" + `
        );
}

function medianCalculator(stats) {
    if (stats.length == 1) {
        return stats[0];
    }
    let rank = 0.5 * (stats.length - 1) + 1;
    let ir = Math.floor(rank);
    let fr = rank - ir;

    return fr * (stats[ir] - stats[ir-1]) + stats[ir-1];
}

function displayBarChart(d, visible) {
  let metric = document.getElementById("metric-selector").value;
  let maxVal = Math.max(...d[metric]);
  let div = d3.select("div.bar-container");
  div.html("");
  if (visible) {
    div
      .style("display", "block")
      .style("top", ` + "`${cursorY - 130}px`" + `)
      .style("left", ` + "`${cursorX}px`" + `)
      .selectAll("div")
      .data(d[metric])
      .enter()
      .append("div")
      .attr("class", "bar")
      .style("height", dd => ` + "`${(dd / maxVal) * 100}%`" + `);
  } else {
    div.style("display", "none");
  }
}

function rowsPopulate(stat) {
    headerPopulate();
    let tBody = d3.select("#trends-table-body");
    tBody.html("");

    tBody
        .selectAll("tr")
        .data(data.results)
        .enter()
	.append("tr")
		.on("mouseover", function(d) {
      		displayBarChart(d, true);
    	})
		.on("mouseout", function(d) {
			displayBarChart(d, false);
		})
        .html(function(d) {
            let row = ` + "`<td title=\"${d.label}\">${d.label}</td>`" + `;
            let validValues = d[stat].filter(val => val != 0);
            validValues.sort((a, b) => a - b);
            let rowMedian = medianCalculator(validValues);
            d[stat].forEach(function(s) {
                if (s == 0) {
                    row += "<td>-</td>";
                } else if (s > rowMedian) {
                    row += ` + "`<td class=\"high\">${s}</td>`" + `;
                } else {
                    row += ` + "`<td>${s}</td>`" + `;
                }
            });

            return row;
        });
}

function resetTable() {
    headerPopulate()
    rowsPopulate(document.getElementById('metric-selector').value)
}

function compare() {
	let testsArray = document.querySelectorAll("#comparison-list input:checked");
	let metric = document.getElementById("metric-selector").value;
	if (testsArray.length < 2) {
		alert("Select at least two tests");
		return;
	}
	let idxs = [];
	testsArray.forEach(function(elem) {
		idxs.push(parseInt(elem.id.slice(3)));
	});
	let headerRow = d3.select("#header-row");
	let tBody = d3.select("#trends-table-body");
	headerRow.html(
		"<th><span>Test Description</span><hr/><span>Request</span></th>"
	);
	tBody.html("");
	headerRow
		.selectAll("th:nth-child(n+2)")
		.data(idxs)
		.enter()
	.append("th")
		.html((d, i) => ` + "`<th onclick=\"sortByColValue(${i})\">${data.tests[d]}</th>`" + `);

	tBody
		.selectAll("tr")
		.data(data.results)
		.enter()
	.append("tr")
		.html(function(d) {
			let row = ` + "`<tr><td title=\"${d.label}\">${d.label}</td>`" + `;
			let vals = d[metric];
			let baselineVal = undefined;
			idxs.forEach(function (idx) {
				let val = vals[idx];
				if (baselineVal === undefined) {
					baselineVal = val;
					if (val == 0) {
						row += "<td>-</td>";
					} else {
						row += ` + "`<td>${val}</td>`" + `;
					}
					return;
				}
				if (val == 0) {
					row += "<td>-</td>";
				} else if (baselineVal == 0 || val == baselineVal) {
					row += ` + "`<td>${val}</td>`" + `;
				} else {
					let diff = Math.round((val / baselineVal - 1) * 100);
					let color = diff > 0 ? "red" : "green";
					row += ` + "`<td>${val} <span style=\"color: ${color}\">(${diff}%)</span></td>`" + `;
				}
			});

			return row;
		})
}

function sortByColValue(idx) {
	let sortClass = document.querySelectorAll("#header-row > th")[idx + 1].classList;
	let allRows = document.querySelectorAll("#trends-table-body > tr");
	let allRowsArray = Array.from(allRows);
	let emptyElems = allRowsArray.filter((elem) => elem.childNodes[idx + 1].innerText.trim() == "-");
	allRowsArray = allRowsArray.filter((elem) => elem.childNodes[idx + 1].innerText.trim().match(/([\d\.]+(\s\(-?\d+%\))?)/));
	let orderDesc = sortClass.contains("desc");
	document.querySelectorAll("#header-row > th:nth-child(n+2)").forEach((elem) => elem.classList.remove("desc"));
	let pattern = /^([\d\.]+)/;
	allRowsArray.sort(function (a, b) {
		let aText = pattern.exec(a.childNodes[idx + 1].innerText)[1];
		let bText = pattern.exec(b.childNodes[idx + 1].innerText)[1];
		let result = 0;
		if (orderDesc) {
			result = aText - bText;
		} else {
			result = bText - aText;
			sortClass.add("desc");
		}

		return result;
	})
	let tableBody = document.querySelector("#trends-table-body");
	tableBody.innerHTML = "";
	allRowsArray.forEach((elem) => tableBody.appendChild(elem));
	emptyElems.forEach((elem) => tableBody.appendChild(elem));
}

window.onload = rowsPopulate("average")

var cursorX;
var cursorY;
document.onmousemove = function(e) {
  cursorX = e.pageX;
  cursorY = e.pageY;
};
`
