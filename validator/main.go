package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"strconv"
	// TODO: Consider removal of Go struct in favor of generic interface for JSON unmarshalling.
	"github.com/ngageoint/seed/objects"
	"github.com/xeipuuv/gojsonschema"
)

type RunResult struct{
	Valid bool
	RunErrors []string
}

func main() {
	schemaUri, dockerImage := GetArgs()

	seedManifest, err := GetSeedManifest(dockerImage)
	checkError(err)

	specResult := ValidateSeedSpec(schemaUri, seedManifest)

	_, nameError := ValidImageName(dockerImage, seedManifest)

	runResult := RunImage(dockerImage, seedManifest)

	DisplayResults(specResult, dockerImage, nameError, runResult)
}

// Validates the Seed manifest in the LABEL of a Docker image against the schema at the specified URI or defaults to the Seed schema on GitHub. 
func ValidateSeedSpec(schemaUri string, seedManifest string) gojsonschema.Result {
	defaultSchemaUri := "https://ngageoint.github.io/seed/schema/seed.manifest.schema.json"
	if (len(schemaUri) == 0) {
		schemaUri = defaultSchemaUri
	}

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
func DockerInspect(dockerImage string) ([]byte, error) {
	cmdstr := "docker"
	args := []string{"inspect", "--format={{json .Config.Labels}}", dockerImage}
	cmd := exec.Command(cmdstr, args...)
	labels, err := cmd.CombinedOutput()
	if err != nil  { //err does not have docker error messages, need to get them from stderr/stdout
		errStr := fmt.Sprintln(err.Error(), string(labels))
		err = errors.New(errStr)  
	} else if strings.HasPrefix(string(labels),"Error:") {
		err = errors.New(string(labels))
	} else if len(labels) == 0 {
		fmt.Println("Empty Docker image label!");
		if err == nil {
			err = errors.New("Docker image has empty label")
		}
	}

	return labels, err
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

func GetSeedManifest(dockerImage string) (string, error) {
	out, err := DockerInspect(dockerImage)
	
	if err != nil {
		return "", err
	}

	seedManifestKey := "com.ngageoint.seed.manifest"
	seedManifest := ParseLabel(out, seedManifestKey)
	if seedManifest == "" {
		err = fmt.Errorf("Seed manifest for image %v is empty!", dockerImage)
	}
	return seedManifest, err
}

// Returns if the given docker image has a valid name that matches this format: <name>-<algorithmVersion>-seed:<packageVersion>
func ValidImageName(dockerImage, seedManifest string) (bool, string) {
	var seed objects.Seed_0_0_3
	err := json.Unmarshal([]byte(seedManifest), &seed)
	if err != nil {
		panic(err.Error())
	}
	temp1 := []string{ seed.Job.Name, seed.Job.AlgorithmVersion, "seed" }
	firstHalf := strings.Join(temp1, "-")
	temp2 := []string{ firstHalf, seed.Job.PackageVersion }
	validString := strings.Join(temp2, ":")

	isNameValid := validString == dockerImage
	nameError := ""
	if !isNameValid {
		str1 := fmt.Sprintln("Docker image name does not match <name>-<algorithmVersion>-seed:<packageVersion> pattern.")
		nameError = fmt.Sprintf("%v Expected %v, given %v\n", str1, validString, dockerImage)
	}
	
	return isNameValid, nameError
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

func RunImage(dockerImage, seedManifest string) RunResult {
//docker run --rm -v "$PWD/testdata":/testdata -e USERID=$UID extractor-0.1.0-seed:0.1.0 testdata/seed-scale.zip -d testdata/
	var result RunResult
	result.Valid = true
	
	var seed objects.Seed_0_0_3
	err := json.Unmarshal([]byte(seedManifest), &seed)
	checkError(err)
	
	inFiles := seed.Job.Interface.InputData.Files
	var volumes []string
	for _, file := range inFiles {
		envName := fmt.Sprintf("${%v}", file.Name)
		filename := os.ExpandEnv(envName)
		if filename == "" {
			if file.Required {
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("%v not defined. Enter filename:", envName)
				filename, _ := reader.ReadString('\n')
				filename = strings.Replace(filename, "\n", "", -1)
				if filename == "" {
					errString := fmt.Sprintf("Required input %v not defined!", envName)
					result.RunErrors = append(result.RunErrors, errString)
					continue
				}
			}
		}
		absPath, _ := filepath.Abs(filename)
		volStr := fmt.Sprintf("%v:/%v", absPath, filename)
		volumes = append(volumes, "-v", volStr)
	}
	
	outFiles := seed.Job.Interface.OutputData.Files
	dockerOutDir := ""
	outAbsPath := ""
	if len(outFiles) > 0 {
		dockerOutDir = os.ExpandEnv("${JOB_OUTPUT_DIR}")
		if dockerOutDir == "" {
			outdir, _ := ioutil.TempDir("", "Output")
			dockerOutDir = filepath.Base(outdir)
			defer os.RemoveAll(outdir) // clean up
			os.Mkdir(dockerOutDir, 0777)
			os.Setenv("${JOB_OUTPUT_DIR}", dockerOutDir)
		}
		outAbsPath, _ = filepath.Abs(dockerOutDir)
		outVolume := outAbsPath + ":/" + dockerOutDir
		volumes = append(volumes, "-v", outVolume)
	}
	
	mounts := seed.Job.Interface.Mounts
	for _, mount := range mounts {
		envName := fmt.Sprintf("${%v}", mount.Name)
		mountHostPath := os.ExpandEnv(envName)
		if mountHostPath == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("%v not defined. Enter the mount path on the host:", envName)
			text, _ := reader.ReadString('\n')
			mountHostPath = strings.Replace(text, "\n", "", -1)
			if mountHostPath == "" {
				errString := fmt.Sprintf("Host path for mount %v not defined!", envName)
				result.RunErrors = append(result.RunErrors, errString)
				continue
			}
		}
		absPath, _ := filepath.Abs(mountHostPath)
		volStr := fmt.Sprintf("%v:/%v", absPath, mount.Path)
		volumes = append(volumes, "-v", volStr)
	}
	
	settings := seed.Job.Interface.Settings
	var envVars []string
	for _, setting := range settings {
		envName := fmt.Sprintf("${%v}", setting.Name)
		setValue := os.ExpandEnv(envName)
		if setValue == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("%v not defined. Enter the %v setting value:", envName, setting.Name)
			setValue, _ := reader.ReadString('\n')
			setValue = strings.Replace(setValue, "\n", "", -1)
			if setValue == "" {
				errString := fmt.Sprintf("Missing value for setting %v", envName)
				result.RunErrors = append(result.RunErrors, errString)
				continue
			}
		}
		envStr := fmt.Sprintf("%v='%v'", setting.Name, setValue)
		envVars = append(envVars, "-e", envStr)
	}

	cmdstr := "docker"
	currentUser, err := user.Current()
	args := []string{"run", "--rm"}
	args = append(args, volumes...)
	args = append(args, envVars...)
	args = append(args, "-u", currentUser.Uid, dockerImage)

	//need to use strings.Replace for output dir in case we set it as env variables set by os.Setenv don't seem to be visible here
	imageArgs := strings.Replace(seed.Job.Interface.Args, "${JOB_OUTPUT_DIR}", dockerOutDir, -1)
	imageArgs = os.ExpandEnv(imageArgs)
	seedArgs := strings.Split(imageArgs, " ")
	args = append(args, seedArgs...)
	
	fmt.Println(args)

	cmd := exec.Command(cmdstr, args...)
	out, err := cmd.CombinedOutput()
	
	if err != nil {
		errString := fmt.Sprintf("Error running image: %v \n %v", string(out), err)
		result.RunErrors = append(result.RunErrors, errString)
	} else {	
		fmt.Printf("Ran image successfully with following output: %v", string(out))
	}
	
	//output files ----------------------
	for _, file := range outFiles {
		if !file.Required {
			continue
		}
		matches, err := filepath.Glob(dockerOutDir + "/" + file.Pattern)
		if err != nil {
			errString := fmt.Sprintf("%v", err)
			result.RunErrors = append(result.RunErrors, errString)
			continue
		}
		count, err := strconv.Atoi(file.Count)
		if err != nil {
			count = 1
		}
		if len(matches) < count {
			format := "Insufficient output. Expected %v files matching pattern %v, found %v"
			errString := fmt.Sprintf(format,  count, file.Pattern, len(matches))
			result.RunErrors = append(result.RunErrors, errString)
		}
	}
	
	//output JSON ------------
	outJson := seed.Job.Interface.OutputData.Json
	documentLoader := gojsonschema.NewStringLoader(string(out))
	processJson := false
	if len(outJson) > 0 {
		_, err := documentLoader.LoadJSON()
		processJson = (err == nil)
		if err != nil {
			stdOutErr := fmt.Sprintf("%v", err)
			manifestPath := filepath.Join(outAbsPath, "results_manifest.json")
			resultsManifest, err := ioutil.ReadFile(manifestPath)
			if err != nil { 
				errString := fmt.Sprintf("Unable to read json from std out: %v", stdOutErr)
				result.RunErrors = append(result.RunErrors, errString)
				errString = fmt.Sprintf("Unable to read results manifest file: %v", manifestPath)
				result.RunErrors = append(result.RunErrors, errString)
			} else {
				documentLoader = gojsonschema.NewStringLoader(string(resultsManifest))
				_, err := documentLoader.LoadJSON()
				processJson = (err == nil)
				if err != nil { 
					errString := fmt.Sprintf("Error loading results manifest file: %v", manifestPath)
					result.RunErrors = append(result.RunErrors, errString)
				}
			}
		}
	}
	
	if processJson {
		schemaFmt := "{ \"type\": \"object\", \"properties\": { %s }, \"required\": [ %s ] }"
		schema := ""
		required := ""
		for i, json := range outJson {
		
			key := json.Name
			if json.Key != "" {
				key = json.Key
			}
		
			schema += fmt.Sprintf("\"%s\": { \"type\": \"%s\" }", key, json.Type)
			if i + 1 < len(outJson) {
				schema += ", "
			}
		
			if json.Required {
				required += fmt.Sprintf("\"%s\",", key)
			}
		}
		required = required[:len(required)-1]
		schema = fmt.Sprintf(schemaFmt, schema, required)

		schemaLoader := gojsonschema.NewStringLoader(schema)
		schemaResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			result.RunErrors = append(result.RunErrors, fmt.Sprintf("Error running validator: %v", err))
		}

		for _, desc := range schemaResult.Errors() {
			result.RunErrors = append(result.RunErrors, fmt.Sprintf("- %s\n", desc))
		}
	
	}
	
	
	result.Valid = (len(result.RunErrors) == 0)
	return result
}

func DisplayResults(result gojsonschema.Result, dockerImage, nameError string, runRes RunResult) {

	if result.Valid() && nameError == "" {
		fmt.Printf("The Docker image %s is valid\n", dockerImage)
	} else {
		fmt.Printf("The Docker image %s is not valid. see errors :\n", dockerImage)
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		fmt.Printf("%s\n", nameError)
	}
	if !runRes.Valid {
		fmt.Printf("There were errors running the docker image:\n")
		for _, desc := range runRes.RunErrors {
			fmt.Printf("- %s\n", desc)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
