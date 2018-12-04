package templates

// JmeterTemplate represents a template for a generated page with Jmeter stats
const JmeterTemplate = `<!DOCTYPE html>
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
tr > td:first-child, tr > th:first-child {
    width: 400px;
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
        headerRow.innerHTML += ` + "`<th onclick=\"sortByColValue(${i})\">${data.tests[i]}</th>`" + `
        comparisonList.innerHTML += ` + "`<li><label for=\"chk${i}\"><input type=\"checkbox\" name=\"chk${i}\" id=\"chk${i}\">${data.tests[i]}</label></li>`" + `
    }
}

function medianCalculator(stats) {
	if (stats.length == 1) {
		return stats[0]
	}
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
        let row = ` + "`<tr><td title=\"${key}\">${key}</td>`" + `
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
		headerRow.innerHTML += ` + "`<th onclick=\"sortByColValue(${i})\">${data.tests[idx]}</th>`" + `
		i++
    })
    let r = data.results
    for (let key in r) {
        let row = ` + "`<tr><td title=\"${key}\">${key}</td>`" + `
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
