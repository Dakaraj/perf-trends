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
	// testing validateParsejmeterArgs function
	if err := validateParseJmeterArgs(parsejmeterCmd, args); err != nil {
		t.Logf("Provided arguments are invalid\nError: %v", err)
		t.Fail()
	}
	// testing parsejmeterFiles function
	parseJmeterFiles(parsejmeterCmd, args)
	os.Remove(args[1])
}
