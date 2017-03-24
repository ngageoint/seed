package main

import (
	"strings"
	"testing"
	"os"
	"path/filepath"
)

func TestValidateSeedSpec(t *testing.T) {
	cases := []struct {
		image string
		expected bool
		expectedErrorMsg string
	}{
		{"seed-test/watermark", true, ""},
		{"seed-test/complete", true, ""},
		{"seed-test/random-number", true, ""},
		{"seed-test/invalid-missing-jobs", false, "jobs: jobs is required"},
		{"seed-test/invalid-missing-job-interface-inputdata-files-name", false, "name: name is required"},
	}
	envSpecUri := os.Getenv("SPEC_PATH")
	absSpecPath := ""
	if (len(envSpecUri) > 0) {
		absPath, _ := filepath.Abs(envSpecUri)
		absSpecPath = filepath.Join("file://", absPath)
	}
	for _, c := range cases {
		result := ValidateSeedSpec(absSpecPath, c.image)
		isValid := result.Valid()
		if (isValid != c.expected ) {
			t.Errorf("ValidateSeedSpec(%q) == %v, expected %v", c.image, isValid, c.expected)
		}
		if (len(result.Errors()) > 0) {
			errorMsg := result.Errors()[0].String()
			if (!strings.Contains(errorMsg, c.expectedErrorMsg)) {
				t.Errorf("Error message contained `%s`, expected `%s`", errorMsg, c.expectedErrorMsg)
			}
		}
	}
}
