/*
Seed implements a command line interface library to build and run
docker images defined by a seed.manifest.json file.
usage is as folllows:
	seed build [OPTIONS]
		Options:
		-d, -directory	The directory containing the seed spec and Dockerfile
										(default is current directory)

	seed list [OPTIONS]
		Not yet implemented

	seed publish [OPTIONS]
		Not yet implemented

	seed run [OPTIONS]
		Options:
		-d, -directory	The directory containing the seed spec and Dockerfile
										(default is current directory)
		-i, -inputData  The input data. May be multiple -id flags defined
										(seedfile: Job.Interface.InputData.Files)
		-in, -imageName The name of the Docker image to run (overrides image name
										pulled from seed spec)
		-o, -outDir			The job output directory. Output defined in
										seedfile: Job.Interface.OutputData.Files and
										Job.Interface.OutputData.Json will be stored relative to
										this directory.
		-rm							Automatically remove the container when it exits (same as
										docker run --rm)
	seed search [OPTIONS]
		Not yet implemented

	seed validate [OPTIONS]
		Options:
			-d, -directory	The directory containing the seed spec
											(default is current directory)
			-s, -schema			Seed Schema file; Overrides built in schema to validate
											spec against.

	seed version
*/
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ngageoint/seed/cli/constants"
	"github.com/ngageoint/seed/cli/objects"
	"github.com/xeipuuv/gojsonschema"
)

var buildCmd *flag.FlagSet
var listCmd *flag.FlagSet
var publishCmd *flag.FlagSet
var runCmd *flag.FlagSet
var searchCmd *flag.FlagSet
var validateCmd *flag.FlagSet
var versionCmd *flag.FlagSet
var directory string
var version string

func main() {

	// Parse input flags
	DefineFlags()

	// seed validate: Validate seed.manifest.json.
	if validateCmd.Parsed() {
		seedFileName := SeedFileName()
		schemaFile := validateCmd.Lookup(constants.SchemaFlag).Value.String()

		if schemaFile != "" {
			schemaFile = "file://" + GetFullPath(schemaFile)
		}

		err := ValidateSeedFile(schemaFile, seedFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
		}
		os.Exit(0)
	}

	// Checks if Docker requires sudo access. Prints error message if so.
	CheckSudo()

	// seed list: Lists all seed compliant images on (default) local machine
	if listCmd.Parsed() {
		DockerList()
		os.Exit(0)
	}

	// seed build: Build Docker image
	if buildCmd.Parsed() {
		DockerBuild("")
		os.Exit(0)
	}

	// seed run: Runs docker image provided or found in seed manifest
	if runCmd.Parsed() {
		DockerRun()
		os.Exit(0)
	}
}

//DockerList lists all seed compliant images (ending with -seed) on the local
//	system
func DockerList() {
	dCmd := exec.Command("docker", "images")
	gCmd := exec.Command("grep", "seed")
	var dErr bytes.Buffer
	dCmd.Stderr = &dErr
	dOut, err := dCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error attaching to std output pipe. %s\n",
			err.Error())
	}

	dCmd.Start()
	if string(dErr.Bytes()) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n",
			string(dErr.Bytes()))
	}

	gCmd.Stdin = dOut
	var gErr bytes.Buffer
	gCmd.Stderr = &gErr

	o, err := gCmd.Output()
	fmt.Fprintf(os.Stderr, string(gErr.Bytes()))
	if string(gErr.Bytes()) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", string(gErr.Bytes()))
	}
	if err != nil && err.Error() != "exit status 1" {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing seed list: %s\n", err.Error())
	}
	if string(o) == "" {
		fmt.Fprintf(os.Stderr, "No Seed Images found!\n")
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", string(o))
	}
}

//DockerBuild Builds the docker image with the given image tag.
func DockerBuild(imageName string) {

	// retrieve seed from seed manifest
	seed, seedFileName := SeedFromManifestFile()

	// Retrieve docker image name
	if imageName == "" {
		imageName = BuildImageName(&seed)
	}

	// Set the seed.manifest.json contents as an image label
	label := "com.ngageoint.seed.manifest=" + GetManifestLabel(seedFileName)

	jobDirectory := buildCmd.Lookup(constants.JobDirectoryFlag).Value.String()

	// Build Docker image
	fmt.Fprintf(os.Stderr, "INFO: Building %s\n", imageName)
	buildArgs := []string{"build", "-t", imageName, jobDirectory}
	if DockerVersionHasLabel() {
		buildArgs = append(buildArgs, "--label", label)
	}
	buildCmd := exec.Command("docker", buildArgs...)

	// attach stderr pipe
	errPipe, err := buildCmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: Error attaching to build command stderr. %s\n", err.Error())
	}

	// Attach stdout pipe
	outPipe, err := buildCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: Error attaching to build command stdout. %s\n", err.Error())
	}

	// Run docker build
	if err := buildCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker build. %s\n",
			err.Error())
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		fmt.Fprintf(os.Stdout, "%s\n", string(slurp))
	}

	// check for errors on stderr
	slurperr, _ := ioutil.ReadAll(errPipe)
	if string(slurperr) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error building image '%s':\n%s\n",
			imageName, string(slurperr))
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}
}

//GetManifestLabel returns the seed.manifest.json as LABEL
//  com.ngageoint.seed.manifest contents
func GetManifestLabel(seedFileName string) string {
	// read the seed.manifest.json into a string
	seedbytes, err := ioutil.ReadFile(seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Eror reading %s. %s\n", seedFileName,
			err.Error())
		os.Exit(1)
	}
	var seedbuff bytes.Buffer
	json.Compact(&seedbuff, seedbytes)
	seedbytes, err = json.Marshal(seedbuff.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error marshalling seed manifest. %s\n",
			err.Error())
	}

	// Escape forward slashes and dollar signs
	seed := string(seedbytes)
	seed = strings.Replace(seed, "$", "\\$", -1)
	seed = strings.Replace(seed, "/", "\\/", -1)

	return seed
}

//GetNormalizedVariable transforms an input name into the spec required environment variable
func GetNormalizedVariable(inputName string) string {
	// Remove all non-alphabetic runes, except dash and underscore
	// Upper-case all lower-case alphabetic runes
	// Dash runes are transformed into underscore
	normalizer := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z' || r == '_':
			return r
		case r >= 'a' && r <= 'z':
			return 'A' + (r - 'a')
		case r == '-':
			return '_'
		}
		return -1
	}

	result := strings.Map(normalizer, inputName)

	return result
}

//SeedFromImageLabel returns seed parsed from the docker image LABEL
func SeedFromImageLabel(imageName string) objects.Seed {
	cmdStr := "inspect -f '{{index .Config.Labels \"com.ngageoint.seed.manifest\"}}'" + imageName
	fmt.Fprintf(os.Stderr,
		"INFO: Retrieving seed manifest from %s LABEL=com.ngageoint.seed.manifest\n",
		imageName)

	inspctCmd := exec.Command("docker", "inspect", "-f",
		"'{{index .Config.Labels \"com.ngageoint.seed.manifest\"}}'", imageName)

	errPipe, errr := inspctCmd.StderrPipe()
	if errr != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: error attaching to docker inspect command stderr. %s\n",
			errr.Error())
	}

	// Attach stdout pipe
	outPipe, errr := inspctCmd.StdoutPipe()
	if errr != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: error attaching to docker inspect command stdout. %s\n",
			errr.Error())
	}

	// Run docker inspect
	if err := inspctCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error executing docker %s. %s\n", cmdStr,
			err.Error())
	}

	// Print out any std out
	seedBytes, err := ioutil.ReadAll(outPipe)
	if err != nil {
		fmt.Fprintf(os.Stdout, "ERROR: Error retrieving docker %s stdout.\n%s\n",
			cmdStr, err.Error())
	}

	// check for errors on stderr
	slurperr, _ := ioutil.ReadAll(errPipe)
	if string(slurperr) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker %s:\n%s\n",
			cmdStr, string(slurperr))
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}

	// un-escape special characters
	seedStr := string(seedBytes)
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1)
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1) //extra replace to fix extra back slashes added by docker build command
	seedStr = strings.Replace(seedStr, "\\$", "$", -1)
	seedStr = strings.Replace(seedStr, "\\/", "/", -1)
	seedStr = strings.TrimSpace(seedStr)
	seedStr = strings.TrimSuffix(strings.TrimPrefix(seedStr, "'\""), "\"'")
	
	seed := &objects.Seed{}

	err = json.Unmarshal([]byte(seedStr), &seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error unmarshalling seed: %s\n", err.Error())
	}
	return *seed
}

//SeedFileName Finds and returns the full filepath to the seed.manifest.json
func SeedFileName() string {
	var dir string

	// Get the proper job directory flag
	if runCmd.Parsed() {
		dir = runCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	} else if buildCmd.Parsed() {
		dir = buildCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	} else if validateCmd.Parsed() {
		dir = validateCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	}

	// Define the current working directory
	curDirectory, _ := os.Getwd()

	// set path to seed file -
	// 	Either relative to current directory or located in given directory
	//  	-d directory might be a relative path to current directory
	seedFileName := constants.SeedFileName
	if dir == "." {
		seedFileName = filepath.Join(curDirectory, seedFileName)
	} else {
		if filepath.IsAbs(dir) {
			seedFileName = filepath.Join(dir, seedFileName)
		} else {
			seedFileName = filepath.Join(curDirectory, dir, seedFileName)
		}
	}

	// Verify seed.json exists within specified directory.
	// If not, error and exit
	if _, err := os.Stat(seedFileName); os.IsNotExist(err) {

		// If no seed.manifest.json found, print the command usage and exit
		if len(os.Args) == 2 {
			PrintCommandUsage()
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s cannot be found. Exiting seed...\n",
				seedFileName)
			os.Exit(1)
		}
	}

	return seedFileName
}

//SeedFromManifestFile returns seed struct parsed from seed file
func SeedFromManifestFile() (objects.Seed, string) {
	seedFileName := SeedFileName()

	// Validate seed file
	err := ValidateSeedFile("", seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: seed file could not be validated. See errors for details.\n")
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}

	// Open and parse seed file into struct
	seedFile, err := os.Open(seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error opening %s. Error received is: %s\n",
			seedFileName, err.Error())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}
	jsonParser := json.NewDecoder(seedFile)
	var seed objects.Seed
	if err = jsonParser.Decode(&seed); err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: A valid %s must be present in the working directory. Error parsing %s.\nError received is: %s\n",
			constants.SeedFileName, seedFileName, err.Error())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}

	return seed, seedFileName
}

//ImageExists returns true if a local image already exists, false otherwise
func ImageExists(imageName string) (bool, error) {
	// Test if image has been built; Rebuild if not
	imgsArgs := []string{"images", "-q", imageName}
	imgOut, err := exec.Command("docker", imgsArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker %v\n", imgsArgs)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return false, err
	} else if string(imgOut) == "" {
		fmt.Fprintf(os.Stderr, "INFO: No docker image found for image name %s. Building image now...\n",
			imageName)
		return false, nil
	}
	return true, nil
}

//DockerRun Runs image described by Seed spec
// func DockerRun(seed *objects.Seed) {
func DockerRun() {
	var seed objects.Seed
	var imageName string

	// Parse seed information off of the label
	if runCmd.Lookup(constants.ImgNameFlag).Value.String() != "" {
		imageName = runCmd.Lookup(constants.ImgNameFlag).Value.String()

		// Check if image exists
		if exists, _ := ImageExists(imageName); !exists {
			// try to build from seed file
			DockerBuild(imageName)
		}

		if runCmd.Lookup(constants.JobDirectoryFlag).Value.String() != "." {
			fmt.Fprintf(os.Stderr,
				"INFO: Image name %s specified. Job directory parameter will be ignored.\n",
				imageName)
		}
		seed = SeedFromImageLabel(imageName)

		// Parse seed from manifest file and build image name
	} else {
		seed, _ = SeedFromManifestFile()
		imageName = BuildImageName(&seed)
		if exists, _ := ImageExists(imageName); !exists {
			DockerBuild(imageName)
		}
	}

	// build docker run command
	dockerArgs := []string{"run"}

	if runCmd.Lookup(constants.RmFlag).Value.String() == "true" {
		dockerArgs = append(dockerArgs, "--rm")
	}

	var mountsArgs []string
	var envArgs []string

	// expand INPUT_FILEs to specified inputData files
	if seed.Job.Interface.InputData.Files != nil {
		inMounts, err := DefineInputs(&seed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error occurred processing inputData arguments.\n%s", err.Error())
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			os.Exit(1)
		} else if inMounts != nil {
			mountsArgs = append(mountsArgs, inMounts...)
		}
	}

	// mount the JOB_OUTPUT_DIR (outDir flag)
	var outDir string
	if strings.Contains(seed.Job.Interface.Cmd, "OUTPUT_DIR") {
		outDir = SetOutputDir(imageName, &seed)
		if outDir != "" {
			mountsArgs = append(mountsArgs, "-v")
			mountsArgs = append(mountsArgs, outDir+":"+outDir)
		}
	}

	// Settings
	if seed.Job.Interface.Settings != nil {
		inSettings, err := DefineSettings(&seed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error occurred processing settings arguments.\n%s", err.Error())
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			os.Exit(1)
		} else if inSettings != nil {
			envArgs = append(envArgs, inSettings...)
		}
	}

	// Additional Mounts defined in seed.json
	if seed.Job.Interface.Mounts != nil {
		inMounts, err := DefineMounts(&seed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error occurred processing mount arguments.\n%s", err.Error())
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			os.Exit(1)
		} else if inMounts != nil {
			mountsArgs = append(mountsArgs, inMounts...)
		}
	}

	// Build Docker command arguments:
	// 		run
	//		-rm if specified
	//		env injection
	// 		all mounts
	//		image name
	//		Job.Interface.Cmd
	dockerArgs = append(dockerArgs, mountsArgs...)
	dockerArgs = append(dockerArgs, envArgs...)
	dockerArgs = append(dockerArgs, imageName)

	// Parse out command arguments from seed.Job.Interface.Cmd
	args := strings.Split(seed.Job.Interface.Cmd, " ")
	dockerArgs = append(dockerArgs, args...)

	// Run
	var cmd bytes.Buffer
	cmd.WriteString("docker ")
	for _, s := range dockerArgs {
		cmd.WriteString(s + " ")
	}
	fmt.Fprintf(os.Stderr, "INFO: Running Docker command:\n%s\n", cmd.String())

	// Run Docker command and capture output
	runCmd := exec.Command("docker", dockerArgs...)
	// attach stderr pipe
	errPipe, err := runCmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error attaching to run command stderr. %s\n",
			err.Error())
	}

	// Attach stdout pipe
	outPipe, err := runCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error attaching to run command stdout. %s\n",
			err.Error())
	}

	// Run docker build
	if err := runCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error executing docker run. %s\n",
			err.Error())
	}

	// check for errors on stderr
	slurperr, _ := ioutil.ReadAll(errPipe)
	if string(slurperr) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error running image '%s':\n%s\n",
			imageName, string(slurperr))
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		fmt.Fprintf(os.Stdout, "%s\n", string(slurp))
	}

	// Validate output against pattern
	if seed.Job.Interface.OutputData.Files != nil ||
		seed.Job.Interface.OutputData.JSON != nil {
		ValidateOutput(&seed, outDir)
	}
}

//DefineFlags defines the flags available for the seed runner.
func DefineFlags() {

	// build command flags
	buildCmd = flag.NewFlagSet(constants.BuildCommand, flag.ContinueOnError)
	buildCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")
	buildCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")

	// Print usage function
	buildCmd.Usage = func() {
		PrintBuildUsage()
	}

	// Run command flags
	runCmd = flag.NewFlagSet(constants.RunCommand, flag.ContinueOnError)
	runCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Location of the seed spec and Dockerfile")
	runCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Location of the seed spec and Dockerfile")

	var imgNameFlag string
	runCmd.StringVar(&imgNameFlag, constants.ImgNameFlag, "",
		"Name of Docker image to run")
	runCmd.StringVar(&imgNameFlag, constants.ShortImgNameFlag, "",
		"Name of Docker image to run")

	var settings objects.ArrayFlags
	runCmd.Var(&settings, constants.SettingFlag,
		"Defines the value to be applied to setting")
	runCmd.Var(&settings, constants.ShortSettingFlag,
		"Defines the value to be applied to setting")

	var mounts objects.ArrayFlags
	runCmd.Var(&mounts, constants.MountFlag,
		"Defines the full path to be mapped via mount")
	runCmd.Var(&mounts, constants.ShortMountFlag,
		"Defines the full path to be mapped via mount")

	var inputs objects.ArrayFlags
	runCmd.Var(&inputs, constants.InputDataFlag,
		"Defines the full path to any input data arguments")
	runCmd.Var(&inputs, constants.ShortInputDataFlag,
		"Defines the full path to input data arguments")

	var outdir string
	runCmd.StringVar(&outdir, constants.JobOutputDirFlag, "",
		"Full path to the algorithm output directory")
	runCmd.StringVar(&outdir, constants.ShortJobOutputDirFlag, "",
		"Full path to the algorithm output directory")

	var rmVar bool
	runCmd.BoolVar(&rmVar, constants.RmFlag, false,
		"Specifying the -rm flag automatically removes the image after executing docker run")

	// Run usage function
	runCmd.Usage = func() {
		PrintRunUsage()
	}

	// List command
	listCmd = flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.Usage = func() {
		PrintListUsage()
	}

	// Search command
	searchCmd = flag.NewFlagSet("search", flag.ExitOnError)
	searchCmd.Usage = func() {
		PrintSearchUsage()
	}

	// Publish command
	publishCmd = flag.NewFlagSet(constants.PublishCommand, flag.ExitOnError)
	publishCmd.Usage = func() {
		PrintPublishUsage()
	}

	// Validate command
	validateCmd = flag.NewFlagSet(constants.ValidateCommand, flag.ExitOnError)
	validateCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Location of the seed.manifest.json spec to validate")
	validateCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Location of the seed.manifest.json spec to validate")
	var schema string
	validateCmd.StringVar(&schema, constants.SchemaFlag, "",
		"JSON schema file to validate seed against.")
	validateCmd.StringVar(&schema, constants.ShortSchemaFlag, "",
		"JSON schema file to validate seed against.")

	validateCmd.Usage = func() {
		PrintValidateUsage()
	}

	// Version command
	versionCmd = flag.NewFlagSet(constants.VersionCommand, flag.ExitOnError)
	versionCmd.Usage = func() {
		PrintVersionUsage()
	}

	if len(os.Args) == 1 {
		PrintUsage()
	}

	switch os.Args[1] {
	case constants.BuildCommand:
		buildCmd.Parse(os.Args[2:])
		if len(buildCmd.Args()) == 1 {
			directory = buildCmd.Args()[0]
		}
	case constants.RunCommand:
		runCmd.Parse(os.Args[2:])
		if len(runCmd.Args()) == 1 {
			directory = runCmd.Args()[0]
		}
	case constants.SearchCommand:
		fmt.Fprintf(os.Stderr, "%q is not yet implemented\n\n", os.Args[1])
		PrintSearchUsage()
	case constants.ListCommand:
		listCmd.Parse(os.Args[2:])
	case constants.PublishCommand:
		fmt.Fprintf(os.Stderr, "%q is not yet implemented\n\n", os.Args[1])
		PrintPublishUsage()
	case constants.ValidateCommand:
		validateCmd.Parse(os.Args[2:])
		if len(validateCmd.Args()) == 1 {
			directory = validateCmd.Args()[0]
		}
	case constants.VersionCommand:
		PrintVersion()
	default:
		fmt.Fprintf(os.Stderr, "%q is not a valid command.\n", os.Args[1])
		PrintUsage()
		os.Exit(0)
	}
}

//PrintCommandUsage prints usage of parsed command, or seed usage if no command
// parsed
func PrintCommandUsage() {
	if buildCmd.Parsed() {
		PrintBuildUsage()
	} else if listCmd.Parsed() {
		PrintListUsage()
	} else if publishCmd.Parsed() {
		PrintPublishUsage()
	} else if runCmd.Parsed() {
		PrintRunUsage()
	} else if searchCmd.Parsed() {
		PrintSearchUsage()
	} else if validateCmd.Parsed() {
		PrintValidateUsage()
	} else {
		PrintUsage()
	}
}

//PrintUsage prints the seed usage arguments
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\tseed COMMAND\n\n")
	fmt.Fprintf(os.Stderr, "A test runner for seed spec compliant algorithms\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  build \tBuilds Seed compliant Docker image\n")
	fmt.Fprintf(os.Stderr, "  list  \tAllows for listing of all Seed compliant images residing on the local system\n")
	fmt.Fprintf(os.Stderr, "  publish\tAllows for publish of Seed compliant images to remote Docker registry\n")
	fmt.Fprintf(os.Stderr, "  run   \tExecutes Seed compliant Docker docker image\n")
	fmt.Fprintf(os.Stderr, "  search\tAllows for discovery of Seed compliant images hosted within a Docker registry (default is docker.io)\n")
	fmt.Fprintf(os.Stderr, "  validate\tValidates a Seed spec\n")
	fmt.Fprintf(os.Stderr, "  version\tPrints the version of Seed spec\n")
	fmt.Fprintf(os.Stderr, "\nRun 'seed COMMAND --help' for more information on a command.\n")
	os.Exit(0)
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintBuildUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\tseed build [-d JOB_DIRECTORY]\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory containing Seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	os.Exit(0)
}

//PrintRunUsage prints the seed run usage arguments, then exits the program
func PrintRunUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed run [-i INPUT_KEY=INPUT_FILE ...] [-e SETTING_KEY=VALUE] -o OUTPUT_DIRECTORY [OPTIONS]\n")
	fmt.Fprintf(os.Stderr, "\nRuns Docker image defined by seed spec.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory containing seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s\tSpecifies the key/value input data values of the seed spec in the format INPUT_FILE_KEY=INPUT_FILE_VALUE\n",
		constants.ShortInputDataFlag, constants.InputDataFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s\tSpecifies the key/value setting values of the seed spec in the format SETTING_KEY=VALUE\n",
		constants.ShortSettingFlag, constants.SettingFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s\tSpecifies the key/value mount values of the seed spec in the format MOUNT_KEY=HOST_PATH\n",
		constants.ShortMountFlag, constants.MountFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tDocker image name to run\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s   \tJob Output Directory Location\n",
		constants.ShortJobOutputDirFlag, constants.JobOutputDirFlag)
	fmt.Fprintf(os.Stderr, "  -%s            \tAutomatically remove the container when it exits (docker run --rm)\n",
		constants.RmFlag)
	os.Exit(0)
}

//PrintListUsage prints the seed list usage information, then exits the program
func PrintListUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed list [OPTIONS]\n")
	fmt.Fprintf(os.Stderr, "\nAllows for listing all Seed compliant images residing on the local system.\n")
	fmt.Fprintf(os.Stderr, "\nLists all '-seed' docker images on the local machine.\n")
	os.Exit(0)
}

//PrintSearchUsage prints the seed search usage information, then exits the program
func PrintSearchUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed search [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nAllows for discovery of seed compliant images hosted within a Docker registry.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -r -repo Specifies a specific registry to search (default is docker.io)\n")
	os.Exit(0)
}

//PrintPublishUsage prints the seed publish usage information, then exits the program
func PrintPublishUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed publish [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nAllows for discovery of seed compliant images hosted within a Docker registry.\n")
	// fmt.Fprintf(os.Stderr, "\nOptions:\n")
	os.Exit(0)
}

//PrintValidateUsage prints the seed validate usage, then exits the program
func PrintValidateUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed validate [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nValidates the given %s is compliant with the Seed spec.\n",
		constants.SeedFileName)
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies directory in which Seed is located (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s   \tExternal Seed schema file; Overrides built in schema to validate Seed spec against\n",
		constants.ShortSchemaFlag, constants.SchemaFlag)
	os.Exit(0)
}

//PrintVersionUsage prints the seed version usage, then exits the program
func PrintVersionUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed version \n")
	fmt.Fprintf(os.Stderr, "\nOutputs the version of the Seed CLI and specification.\n")
	os.Exit(0)
}

//PrintVersion prints the seed CLI version
func PrintVersion() {
	fmt.Fprintf(os.Stderr, "Seed v%s\n", version)
	os.Exit(0)
}

//BuildImageName extracts the Docker Image name from either the input arguments
// (via commnad flags -jobName -algVersion -pkgVersion) or the seed.json, or
// combination of th etwo. Returns image name in the form of
// 	jobName-algVersion-seed:pkgVersion
func BuildImageName(seed *objects.Seed) string {
	var buffer bytes.Buffer

	if runCmd.Parsed() && runCmd.Lookup(constants.ImgNameFlag).Value.String() != "" {
		buffer.WriteString(runCmd.Lookup(constants.ImgNameFlag).Value.String())
	} else {
		buffer.WriteString(seed.Job.Name)
		buffer.WriteString("-")
		buffer.WriteString(seed.Job.AlgorithmVersion)
		buffer.WriteString("-seed")
		buffer.WriteString(":")
		buffer.WriteString(seed.Job.PackageVersion)
	}

	return buffer.String()
}

//DefineInputs extracts the paths to any input data given by the 'run' command
// flags 'inputData' and sets the path in the json object. Returns:
// 	[]string: docker command args for input files in the format:
//	"-v /path/to/file1:/path/to/file1 -v /path/to/file2:/path/to/file2 etc"
func DefineInputs(seed *objects.Seed) ([]string, error) {

	// Validate inputs given vs. inputs defined in manifest

	// Ingest inputs into a map key = inputkey, value=inputpath
	inputs := strings.Split(runCmd.Lookup(constants.InputDataFlag).Value.String(), ",")
	inMap := make(map[string]string)
	for _, f := range inputs {
		x := strings.Split(f, "=")
		if len(x) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: Input files should be specified in KEY=VALUE format.\n")
			fmt.Fprintf(os.Stderr, "ERROR: Unknown key for input %v encountered.\n",
				x)
			continue
		}
		inMap[x[0]] = x[1]
	}

	// Valid by default
	valid := true
	var keys []string
	for _, f := range seed.Job.Interface.InputData.Files {
		keys = append(keys, f.Name)
		if _, prs := inMap[f.Name]; !prs {
			valid = false
		}
	}

	if !valid {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR: Incorrect input data files key/values provided. -i arguments should be in the form:\n")
		buffer.WriteString("  seed run -i KEY1=path/to/file1 -i KEY2=path/to/file2 ...\n")
		buffer.WriteString("The following input file keys are expected:\n")
		for _, n := range keys {
			buffer.WriteString("  " + n + "\n")
		}
		buffer.WriteString("\n")
		return nil, errors.New(buffer.String())
	}

	// TODO: validate number of inputData flags to number of Interface.InputData.Files
	var mountArgs []string

	for _, f := range inputs {
		x := strings.Split(f, "=")
		if len(x) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: Input files should be specified in KEY=VALUE format.\n")
			fmt.Fprintf(os.Stderr, "ERROR: Unknown key for input %v encountered.\n",
				inputs)
			continue
		}

		key := x[0]
		val := x[1]

		// Expand input VALUE
		val = GetFullPath(val)

		// Replace key if found in args strings
		// Handle replacing KEY or ${KEY} or $KEY
		seed.Job.Interface.Cmd = strings.Replace(seed.Job.Interface.Cmd,
			"${"+key+"}", val, -1)
		seed.Job.Interface.Cmd = strings.Replace(seed.Job.Interface.Cmd, "$"+key,
			val, -1)
		seed.Job.Interface.Cmd = strings.Replace(seed.Job.Interface.Cmd, key, val,
			-1)

		for _, k := range seed.Job.Interface.InputData.Files {
			if k.Name == key {
				mountArgs = append(mountArgs, "-v")
				mountArgs = append(mountArgs, val+":"+val)
			}
		}
	}

	return mountArgs, nil
}

//SetOutputDir replaces the OUTPUT_DIR argument with the given output directory.
// Returns output directory string
func SetOutputDir(imageName string, seed *objects.Seed) string {

	if !strings.Contains(seed.Job.Interface.Cmd, "OUTPUT_DIR") {
		return ""
	}

	outputDir := runCmd.Lookup(constants.JobOutputDirFlag).Value.String()

	// #37: if -o is not specified, and OUTPUT_DIR is in the command args,
	//	auto create a time-stamped subdirectory with the name of the form:
	//		imagename-iso8601timestamp
	if outputDir == "" {
		outputDir = "output-" + imageName + "-" + time.Now().Format(time.RFC3339)
		outputDir = strings.Replace(outputDir, ":", "_", -1)
	}

	outdir := GetFullPath(outputDir)

	// Check if outputDir exists. Create if not
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		// Create the directory
		// Didn't find the specified directory
		fmt.Fprintf(os.Stderr, "INFO: %s not found; creating directory...\n",
			outdir)
		os.Mkdir(outdir, os.ModePerm)
	}

	// Check if outdir is empty. Create time-stamped subdir if not
	f, err := os.Open(outdir)
	if err != nil {
		// complain
		fmt.Fprintf(os.Stderr, "ERROR: Error with %s. %s\n", outdir, err.Error())
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err != io.EOF {
		// Directory is not empty
		t := time.Now().Format("20060102_150405")
		fmt.Fprintf(os.Stderr,
			"INFO: Output directory %s is not empty. Creating sub-directory %s for Job Output Directory.\n",
			outdir, t)
		outdir = filepath.Join(outdir, t)
		os.Mkdir(outdir, os.ModePerm)
	}

	seed.Job.Interface.Cmd = strings.Replace(seed.Job.Interface.Cmd,
		"$OUTPUT_DIR", outdir, -1)
	seed.Job.Interface.Cmd = strings.Replace(seed.Job.Interface.Cmd,
		"${OUTPUT_DIR}", outdir, -1)
	return outdir
}

//DefineMounts defines any seed specified mounts. TODO
func DefineMounts(seed *objects.Seed) ([]string, error) {
	// Ingest inputs into a map key = inputkey, value=inputpath
	inputs := strings.Split(runCmd.Lookup(constants.MountFlag).Value.String(), ",")
	inMap := make(map[string]string)
	for _, f := range inputs {
		x := strings.Split(f, "=")
		if len(x) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: Mount should be specified in KEY=VALUE format.\n")
			fmt.Fprintf(os.Stderr, "ERROR: Unknown key for mount %v encountered.\n",
				x)
			continue
		}
		inMap[x[0]] = x[1]
	}

	// Valid by default
	valid := true
	var keys []string
	for _, f := range seed.Job.Interface.Mounts {
		keys = append(keys, f.Name)
		if _, prs := inMap[f.Name]; !prs {
			valid = false
		}
	}

	if !valid {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR: Incorrect mount key/values provided. -m arguments should be in the form:\n")
		buffer.WriteString("  seed run -m MOUNT=path/to ...\n")
		buffer.WriteString("The following mount keys are expected:\n")
		for _, n := range keys {
			buffer.WriteString("  " + n + "\n")
		}
		buffer.WriteString("\n")
		return nil, errors.New(buffer.String())
	}

	var mounts []string
	if seed.Job.Interface.Mounts != nil {
		for _, mount := range seed.Job.Interface.Mounts {
			mounts = append(mounts, "-v")
			mountPath := inMap[mount.Name] + ":" + mount.Path

			if mount.Mode != "" {
				mountPath += ":" + mount.Mode
			} else {
				mountPath += ":ro"
			}
			mounts = append(mounts, mountPath)
		}
		return mounts, nil
	}

	return mounts, nil
}

//DefineSettings defines any seed specified docker settings. TODO
// Return []string of docker command arguments in form of:
//	"-?? setting1=val1 -?? setting2=val2 etc"
func DefineSettings(seed *objects.Seed) ([]string, error) {
	// Ingest inputs into a map key = inputkey, value=inputpath
	inputs := strings.Split(runCmd.Lookup(constants.SettingFlag).Value.String(), ",")
	inMap := make(map[string]string)
	for _, f := range inputs {
		x := strings.Split(f, "=")
		if len(x) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: Setting should be specified in KEY=VALUE format.\n")
			fmt.Fprintf(os.Stderr, "ERROR: Unknown key for setting %v encountered.\n",
				x)
			continue
		}
		inMap[x[0]] = x[1]
	}

	// Valid by default
	valid := true
	var keys []string
	for _, s := range seed.Job.Interface.Settings {
		keys = append(keys, s.Name)
		if _, prs := inMap[s.Name]; !prs {
			valid = false
		}

	}

	if !valid {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR: Incorrect setting key/values provided. -e arguments should be in the form:\n")
		buffer.WriteString("  seed run -e SETTING=somevalue ...\n")
		buffer.WriteString("The following settings are expected:\n")
		for _, n := range keys {
			buffer.WriteString("  " + n + "\n")
		}
		buffer.WriteString("\n")
		return nil, errors.New(buffer.String())
	}

	var settings []string
	for _, key := range keys {
		settings = append(settings, "-e")
		settings = append(settings, GetNormalizedVariable(key)+"="+inMap[key])
	}

	return settings, nil
}

//ValidateOutput validates the output of the docker run command. Output data is
// validated as defined in the seed.Job.Interface.OutputData.
func ValidateOutput(seed *objects.Seed, outDir string) {
	// Validate any OutputData.Files
	if seed.Job.Interface.OutputData.Files != nil {
		fmt.Fprintf(os.Stderr, "INFO: Validating output files found under %s...\n",
			outDir)

		// For each defined OutputData file:
		//	#1 Check file media type
		// 	#2 Check file names match output pattern
		//  #3 Check number of files (if defined)
		for _, f := range seed.Job.Interface.OutputData.Files {

			// find all pattern matches in OUTPUT_DIR
			matches, _ := filepath.Glob(path.Join(outDir, f.Pattern))

			// Check media type of matches
			count := 0
			var matchList []string
			for _, match := range matches {
				ext := filepath.Ext(match)
				mType := mime.TypeByExtension(ext)
				if strings.Contains(mType, f.MediaType) ||
					strings.Contains(f.MediaType, mType) {
					count++
					matchList = append(matchList, "\t"+match+"\n")
				}
			}

			// Validate number of matches to specified number
			if f.Count != "" && f.Count != "*" {
				count, _ := strconv.Atoi(f.Count)
				if count != len(matchList) {
					fmt.Fprintf(os.Stderr, "ERROR: %v files specified, %v found.\n",
						f.Count, strconv.Itoa(len(matchList)))
					if len(matchList) > 0 {
						for _, s := range matchList {
							fmt.Fprintf(os.Stderr, s)
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "SUCCESS: %v files specified, %v found. Files found:\n",
						f.Count, strconv.Itoa(len(matchList)))
					for _, s := range matchList {
						fmt.Fprintf(os.Stderr, s)
					}
				}
			}
		}
	}

	// Validate any defined OutputData.Json
	// Look for ResultsFileManifestName.json in the root of the OUTPUT_DIR
	// and then validate any keys identified in OutputData exist
	if seed.Job.Interface.OutputData.JSON != nil {
		fmt.Fprintf(os.Stderr, "INFO: Validating %s...\n",
			filepath.Join(outDir, constants.ResultsFileManifestName))
		// look for results manifest
		manfile := filepath.Join(outDir, constants.ResultsFileManifestName)
		if _, err := os.Stat(manfile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %s specified but cannot be found. %s\n Exiting testrunner.\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}

		bites, err := ioutil.ReadFile(filepath.Join(outDir,
			constants.ResultsFileManifestName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error reading %s.%s\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}
		
		documentLoader := gojsonschema.NewStringLoader(string(bites))
		_, err = documentLoader.LoadJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error loading results manifest file: %s. %s\n Exiting testrunner.\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}
		
		schemaFmt := "{ \"type\": \"object\", \"properties\": { %s }, \"required\": [ %s ] }"
		schema := ""
		required := ""

		// Loop through defined name/key values to extract from results_manifest.json
		for _, jsonStr := range seed.Job.Interface.OutputData.JSON {
			key := jsonStr.Name
			if jsonStr.Key != "" {
				key = jsonStr.Key
			}

			schema += fmt.Sprintf("\"%s\": { \"type\": \"%s\" },", key, jsonStr.Type)

			if jsonStr.Required {
				required += fmt.Sprintf("\"%s\",", key)
			}
		}
		//remove trailing commas
		if len(schema) > 0 {
			schema = schema[:len(schema)-1]
		}
		if len(required) > 0 {
			required = required[:len(required)-1]
		}
		
		schema = fmt.Sprintf(schemaFmt, schema, required)
		
		schemaLoader := gojsonschema.NewStringLoader(schema)
		schemaResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error running validator: %s\n Exiting testrunner.\n",
				err.Error())
			return
		}
		
		if len(schemaResult.Errors()) == 0 {
			fmt.Fprintf(os.Stderr, "SUCCESS: Results manifest file is valid.\n")
		}

		for _, desc := range schemaResult.Errors() {
			fmt.Fprintf(os.Stderr, "ERROR: %s is invalid: - %s\n", constants.ResultsFileManifestName, desc)
		}
	}
}

//ValidateSeedFile Validates the seed.manifest.json file based on the given schema
func ValidateSeedFile(schemaFile string, seedFileName string) error {
	var result *gojsonschema.Result
	var err error

	// Load supplied schema file
	if schemaFile != "" {
		fmt.Fprintf(os.Stderr, "INFO: Validating seed file %s against schema file %s...\n",
			seedFileName, schemaFile)
		schemaLoader := gojsonschema.NewReferenceLoader(schemaFile)
		docLoader := gojsonschema.NewReferenceLoader("file://" + seedFileName)
		result, err = gojsonschema.Validate(schemaLoader, docLoader)

		// Load baked-in schema file
	} else {
		fmt.Fprintf(os.Stderr, "INFO: Validating seed file %s against schema...\n",
			seedFileName)
		schemaBytes, _ := constants.Asset("../spec/schema/seed.manifest.schema.json")
		schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
		docLoader := gojsonschema.NewReferenceLoader("file://" + seedFileName)
		result, err = gojsonschema.Validate(schemaLoader, docLoader)
	}

	// Error occurred loading the schema or seed.manifest.json
	if err != nil {
		// fmt.Fprintf(os.Stderr,
		// 	"ERROR: Error validating seed file against schema. Error is: %s\n",
		// 	err.Error())
		return errors.New("ERROR: Error validating seed file against schema. Error is:" + err.Error() + "\n")
	}

	// Validation failed. Print results
	if !result.Valid() {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR:" + seedFileName + " is not valid. See errors:\n")
		for _, e := range result.Errors() {
			buffer.WriteString("-ERROR " + e.Description() + "\n")
			buffer.WriteString("\tField: " + e.Field() + "\n")
			buffer.WriteString("\tContext: " + e.Context().String() + "\n")
		}
		return errors.New(buffer.String())
	}

	// TODO Identify any name collisions
	// searches through the job.interface.inputData.Files/JSON and interface.settings
	// for the follwing reserved variables:
	//		OUTPUT_DIR, ALLOCATED_CPUS, ALLOCATED_MEM, ALLOCATED_SHARED_MEM, ALLOCATED_STORAGE

	// Validation succeeded
	fmt.Fprintf(os.Stderr, "SUCCESS: %s is valid.\n\n", seedFileName)
	return nil
}

//GetFullPath returns the full path of the given file. This expands relative file
// paths and verifes non-relative paths
// Validate path for file existance??
func GetFullPath(rFile string) string {

	// Normalize
	rFile = filepath.Clean(filepath.ToSlash(rFile))

	if !filepath.IsAbs(rFile) {

		// Expand relative file path
		// Define the current working directory
		curDir, _ := os.Getwd()

		// Test relative to current directory
		dir := filepath.Join(curDir, rFile)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			rFile = filepath.Clean(dir)

			// see if parent directory exists. If so, assume this directory will be created
		} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
			rFile = filepath.Clean(dir)
		}

		// Test relative to working directory
		if directory != "." {
			dir = filepath.Join(directory, rFile)
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)

				// see if parent directory exists. If so, assume this directory will be created
			} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)
			}
		}
	}

	return rFile
}

//CheckSudo Checks error for telltale sign seed command should be run as sudo
func CheckSudo() {
	cmd := exec.Command("docker", "info")

	// attach stderr pipe
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: Error attaching to version command stderr. %s\n", err.Error())
	}

	// Run docker build
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker version. %s\n",
			err.Error())
	}

	slurperr, _ := ioutil.ReadAll(errPipe)
	er := string(slurperr)
	if er != "" {
		if strings.Contains(er, "Cannot connect to the Docker daemon. Is the docker daemon running on this host?") ||
			strings.Contains(er, "dial unix /var/run/docker.sock: connect: permission deied") {
			fmt.Fprintf(os.Stderr, "Elevated permissions are required by seed to run Docker. Try running the seed command again as sudo.\n")
		}
		os.Exit(1)
	}
}

//DockerVersionHasLabel returns if the docker version is greater than 1.11.1
func DockerVersionHasLabel() bool {
	cmd := exec.Command("docker", "--version")

	// Attach stdout pipe
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: Error attaching to version command stdout. %s\n", err.Error())
	}

	// Run docker version
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker version. %s\n",
			err.Error())
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		// Trim extra text
		versionStr := strings.TrimPrefix(string(slurp), "Docker version ")
		versionStr = versionStr[0:strings.Index(versionStr, "-")]
		version := strings.Split(versionStr, ".")

		// check each part of version. Return false if 1st < 1, 2nd < 11, 3rd < 1
		if len(version) > 1 {
			v1, _ := strconv.Atoi(version[0])
			v2, _ := strconv.Atoi(version[1])

			// check for minimum of 1.11.1
			if v1 == 1 {
				if v2 > 11 {
					return true
				} else if v2 == 11 && len(version) == 3 {
					v3, _ := strconv.Atoi(version[2])
					if v3 >= 1 {
						return true
					}
				}
			} else if v1 > 1 {
				return true
			}

			return false
		}
	}

	return false
}
