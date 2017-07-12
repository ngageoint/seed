package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSeedSpec(t *testing.T) {
	cases := []struct {
		image            string
		expected         bool
		expectedErrorMsg string
	}{
		{"image-watermark-0.1.0-seed:0.1.0", true, ""},
		{"my-algorithm-0.1.0-seed:0.1.0", true, ""},
		{"random-number-gen-0.1.0-seed:0.1.0", true, ""},
		{"extractor-0.1.0-seed:0.1.0", true, ""},
		{"seed-test/invalid-missing-job", false, "job: job is required"},
		{"missing-filename-0.1.0-seed:0.1.0", false, "name: name is required"},
	}
	envSpecUri := os.Getenv("SPEC_PATH")
	absSpecPath := ""
	if len(envSpecUri) > 0 {
		absPath, _ := filepath.Abs(envSpecUri)
		absSpecPath = filepath.Join("file://", absPath)
	}

	for _, c := range cases {
		seedManifest, err := GetSeedManifest(c.image)
		if err != nil {
			t.Errorf("Error getting manifest: %v", err)
		}
		result := ValidateSeedSpec(absSpecPath, seedManifest)
		isValid := result.Valid()
		if isValid != c.expected {
			t.Errorf("ValidateSeedSpec(%q) == %v, expected %v", c.image, isValid, c.expected)
		}
		if len(result.Errors()) > 0 {
			errorMsg := result.Errors()[0].String()
			if !strings.Contains(errorMsg, c.expectedErrorMsg) {
				t.Errorf("Error message contained `%s`, expected `%s`", errorMsg, c.expectedErrorMsg)
			}
		}
	}
}

func TestValidImageName(t *testing.T) {
	cases := []struct {
		image            string
		valid            bool
		expectedErrorMsg string
	}{
		{"image-watermark-0.1.0-seed:0.1.0", true, ""},
		{"my-algorithm-0.1.0-seed:0.1.0", true, ""},
		{"extractor-0.1.0-seed:0.1.0", true, ""},
		{"seed-test/invalid-missing-job", false, "Expected --seed:, given seed-test/invalid-missing-job"},
		{"seed-test/watermark", false, "Expected image-watermark-0.1.0-seed:0.1.0, given seed-test/watermark"},
	}
	for _, c := range cases {
		seedManifest, err := GetSeedManifest(c.image)
		valid, errorMsg := ValidImageName(c.image, seedManifest)
		if err != nil {
			continue
		}
		if valid != c.valid {
			t.Errorf("ValidImageName(%q) == %v, expected %v", c.image, valid, c.valid)
		}
		if (len(errorMsg) > 0) && (!strings.Contains(errorMsg, c.expectedErrorMsg)) {
			t.Errorf("Error message contained `%s`, expected `%s`", errorMsg, c.expectedErrorMsg)
		}
	}
}

func TestRunImage(t *testing.T) {
	cases := []struct {
		image            string
		valid            bool
		expectedErrorMsg string
	}{
		{"extractor-0.1.0-seed:0.1.0", true, ""},
	}
	for _, c := range cases {
		seedManifest, err := GetSeedManifest(c.image)
		if err != nil {
			continue
		}
		result := RunImage(c.image, seedManifest)
		if result.Valid != c.valid {
			t.Errorf("RunImage(%q) == %v, expected %v", c.image, result.Valid, c.valid)
		}
		if len(result.RunErrors) > 0 {
			errorMsg := result.RunErrors[0]
			if !strings.Contains(errorMsg, c.expectedErrorMsg) {
				t.Errorf("Error message contained `%s`, expected `%s`", errorMsg, c.expectedErrorMsg)
			}
		}
	}
}
