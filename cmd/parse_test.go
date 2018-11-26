package cmd

import (
	"os"
	"testing"
)

func TestParsingJmeterLog(t *testing.T) {
	args := []string{
		"SomeTest",
		"./testdata/somefile.db",
		"./testdata/test-input.csv",
	}
	header = true
	delimiter = "~"
	ignorePatternString = `^(TC |OPTIONS |chunk\.)`
	// testing validateParseArgs function
	if err := validateParseArgs(parseCmd, args); err != nil {
		t.Logf("Provided arguments are invalid\nError: %v", err)
		t.Fail()
	}
	// testing parseFiles function
	parseFiles(parseCmd, args)
	os.Remove(args[1])
}
