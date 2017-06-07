package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"seed/objects"
	"github.com/xeipuuv/gojsonschema"
)

type NameError struct{
}

func main() {
	schemaUri, dockerImage := GetArgs()
	result, validImageName := ValidateSeedSpec(schemaUri, dockerImage)
	DisplayResults(result, dockerImage, validImageName)
}

// Validates the Seed manifest in the LABEL of a Docker image against the schema at the specified URI or defaults to the Seed schema on GitHub. 
func ValidateSeedSpec(schemaUri string, dockerImage string) (gojsonschema.Result, string) {
	defaultSchemaUri := "https://ngageoint.github.io/seed/schema/seed.manifest.schema.json"
	seedManifestKey := "com.ngageoint.seed.manifest"
	if (len(schemaUri) == 0) {
		schemaUri = defaultSchemaUri
	}

	out, err := DockerInspect(dockerImage)
	checkError(err)
	seedManifest := ParseLabel(out, seedManifestKey)
	result := Validate(schemaUri, seedManifest)
	validName := ValidImageName(seedManifest)

	return result, validName
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
func DockerInspect(dockerImage string) ([]byte, error) {
	cmdstr := "docker"
	args := []string{"inspect", "--format={{json .Config.Labels}}", dockerImage}
	cmd := exec.Command(cmdstr, args...)
	out, err := cmd.CombinedOutput()
	if err != nil  { //err does not have docker error messages, need to get them from stderr/stdout
		errStr := fmt.Sprintln(err.Error(), string(out))
		err = errors.New(errStr)  
	} else if strings.HasPrefix(string(out),"Error:") {
		err = errors.New(string(out))
	} else if len(out) == 0 {
		fmt.Println("Empty Docker image label!");
		if err == nil {
			err = errors.New("Docker image has empty label")
		}
	}
	return out, err
}

// Convert stdout to string, parse into a map, and retrieve Seed info by key
func ParseLabel(stdout []byte, seedManifestKey string) string {
	outputJson := string(stdout)
	labelMap := make(map[string]string)
	err := json.Unmarshal([]byte(outputJson), &labelMap)
	if err != nil {
		checkError(err)
	}
	seedManifest, ok := labelMap[seedManifestKey]
	if !ok {
		err = errors.New("Docker image label is missing seed manifest key")
		checkError(err)
	}

	return seedManifest
}

// Returns a valid image name in this format from a given seed manifest: <name>-<algorithmVersion>-seed:<packageVersion>
func ValidImageName(seedManifest string) string {
	var seed objects.Seed_0_0_3
	err := json.Unmarshal([]byte(seedManifest), &seed)
	if err != nil {
		panic(err.Error())
	}
	temp1 := []string{ seed.Job.Name, seed.Job.AlgorithmVersion, "seed" }
	firstHalf := strings.Join(temp1, "-")
	temp2 := []string{ firstHalf, seed.Job.PackageVersion }
	validString := strings.Join(temp2, ":")
	
	return validString
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

func DisplayResults(result gojsonschema.Result, dockerImage string, validName string) {
	isNameValid := validName == dockerImage
	var nameError string
	if !isNameValid {
		str1 := fmt.Sprintln("Docker image name does not match <name>-<algorithmVersion>-seed:<packageVersion> pattern.")
		nameError = fmt.Sprintf("%s Expected %s, given %s\n", str1, validName, dockerImage)
	}
	if result.Valid() && isNameValid {
		fmt.Printf("The Docker image %s is valid\n", dockerImage)
	} else {
		fmt.Printf("The Docker image %s is not valid. see errors :\n", dockerImage)
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		fmt.Printf("%s\n", nameError)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
