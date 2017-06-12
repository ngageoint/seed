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
		{"image-watermark-0.1.0-seed:0.1.0", true, ""},
		{"my-algorithm-0.1.0-seed:0.1.0", true, ""},
		{"random-number-gen-0.1.0-seed:0.1.0", true, ""},
		{"seed-test/invalid-missing-job", false, "job: job is required"},
		{"missing-filename-0.1.0-seed:0.1.0", false, "name: name is required"},
	}
	envSpecUri := os.Getenv("SPEC_PATH")
	absSpecPath := ""
	if (len(envSpecUri) > 0) {
		absPath, _ := filepath.Abs(envSpecUri)
		absSpecPath = filepath.Join("file://", absPath)
	}
	for _, c := range cases {
		result, _ := ValidateSeedSpec(absSpecPath, c.image)
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

func TestValidImageName(t *testing.T) {
	cases := []struct {
		image string
		expected string
	}{
		{"image-watermark-0.1.0-seed:0.1.0", "image-watermark-0.1.0-seed:0.1.0"},
		{"my-algorithm-0.1.0-seed:0.1.0", "my-algorithm-0.1.0-seed:0.1.0"},
		{"seed-test/invalid-missing-job", "--seed:"},
		{"seed-test/watermark", "image-watermark-0.1.0-seed:0.1.0" },
	}
	for _, c := range cases {
		out, err := DockerInspect(c.image)
		seedManifest := ParseLabel(out, "com.ngageoint.seed.manifest")
		result := ValidImageName(string(seedManifest))
		if err != nil {
			continue
		}
		if (result != c.expected ) {
			t.Errorf("ValidImageName(%q) == %v, expected %v", c.image, result, c.expected)
		}
	}
}
