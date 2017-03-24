package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"github.com/xeipuuv/gojsonschema"
)

func main() {
	schemaUri, dockerImage := GetArgs()
	result := ValidateSeedSpec(schemaUri, dockerImage)
	DisplayResults(result, dockerImage)
}

// Validates the Seed manifest in the LABEL of a Docker image against the schema at the specified URI or defaults to the Seed schema on GitHub. 
func ValidateSeedSpec(schemaUri string, dockerImage string) gojsonschema.Result {
	defaultSchemaUri := "https://ngageoint.github.io/seed/schema/seed.manifest.schema.json"
	seedManifestKey := "com.ngageoint.seed.manifest"
	if (len(schemaUri) == 0) {
		schemaUri = defaultSchemaUri
	}

	out := DockerInspect(dockerImage)
	seedManifest := ParseLabel(out, seedManifestKey)
	result := Validate(schemaUri, seedManifest)
	return result
}

// Retrieve command line arguments
func GetArgs() (string, string) {
	var schemaUri string
	var dockerImage string

	flag.StringVar(&schemaUri, "schema", "", "An optional URI of a schema to validate against.")
	flag.StringVar(&dockerImage, "image", "", "A Docker image to validate against the Seed specification.")
	flag.Parse()
	if (len(dockerImage) == 0) {
		fmt.Println("\n\"seedvalidator\" requires a docker image be specified.\n\nUsage: seedvalidator -image DOCKERIMAGE \n")
		os.Exit(1)
	}
	return schemaUri, dockerImage
}

// Execute 'docker inspect' to parse Dockerfile and get json LABEL info from stdout
func DockerInspect(dockerImage string) []byte {
	cmd := "docker"
	args := []string{"inspect", "--format={{json .Config.Labels}}", dockerImage}
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Println("An error occurred!")
		fmt.Fprintln(os.Stderr, err)
	}
	// TODO: test for no output from docker inspect
	return out
}

// Convert stdout to string, parse into a map, and retrieve Seed info by key
func ParseLabel(stdout []byte, seedManifestKey string) string {
	outputJson := string(stdout)
	labelMap := make(map[string]string)
	err := json.Unmarshal([]byte(outputJson), &labelMap)
	if err != nil {
		panic(err.Error())
	}
	seedManifest := labelMap[seedManifestKey]
	// TODO: test for seed manifest key not found
	return seedManifest
}

// Load JSON manifest and validate against Seed schema
func Validate(schemaUri string, seedManifest string) gojsonschema.Result {
	schemaLoader := gojsonschema.NewReferenceLoader(schemaUri)
	documentLoader := gojsonschema.NewStringLoader(seedManifest)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}
	return *result
}

func DisplayResults(result gojsonschema.Result, dockerImage string) {
	if result.Valid() {
		fmt.Printf("The Docker image %s is valid\n", dockerImage)
	} else {
		fmt.Printf("The Docker image %s is not valid. see errors :\n", dockerImage)
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}
