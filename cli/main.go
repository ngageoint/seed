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
*/
package main

import (
	"bytes"
	"encoding/json"
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

	"./constants"
	"./objects"
	"github.com/xeipuuv/gojsonschema"
)

var seedFileName string
var buildCmd *flag.FlagSet
var listCmd *flag.FlagSet
var publishCmd *flag.FlagSet
var runCmd *flag.FlagSet
var searchCmd *flag.FlagSet
var validateCmd *flag.FlagSet
var directory string
var curDirectory string

/* Run command defaults:
   image name and tag: if not specified, attempt to guess from CWD if a
      seed.json exists. Otherwise error.
   input: no default; args should match file InputData as described in
      seed.json. You'll need to search / replace this with container
      resolvable paths. It's the algorithm developers responsibility to
      create parameter expansion
   output: no default; single directory where output files are placed. Glob
      capture expressions are described in seed.json
*/
func main() {

	// Parse input flags
	DefineFlags()

	// Lists all seed compliant images on (default) local machine
	if listCmd.Parsed() {
		DockerList()
		os.Exit(1)
	}

	// Define the current working directory
	curDirectory, _ := os.Getwd()

	// set path to seed file -
	// 	Either relative to current directory or located in given directory (-d option)
	//  	-d directory might be a relative path to current directory
	seedFileName = constants.SeedFileName
	if directory == "." {
		directory = curDirectory
		seedFileName = filepath.Join(curDirectory, seedFileName)
	} else {
		if filepath.IsAbs(directory) {
			seedFileName = filepath.Join(directory, seedFileName)
		} else {
			seedFileName = filepath.Join(curDirectory, directory, seedFileName)
		}
	}

	// Verify seed.json exists within specified directory.
	// If not, error and exit
	if _, err := os.Stat(seedFileName); os.IsNotExist(err) {

		// If no seed.manifest.json found, print the command usage and exit
		if len(os.Args) == 2 {
			PrintCommandUsage()
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s cannot be found. Exiting seed...\n",
				seedFileName)
			os.Exit(1)
		}
	}

	// Validate seed.json. Exit if invalid
	if validateCmd.Parsed() {
		schemaFile := validateCmd.Lookup(constants.SchemaFlag).Value.String()

		if schemaFile != "" {
			schemaFile = "file://" + GetFullPath(schemaFile)
		}

		ValidateSeedFile(schemaFile, seedFileName)
		os.Exit(0)
	} else {
		valid := ValidateSeedFile("", seedFileName)
		if !valid {
			fmt.Fprintf(os.Stderr, "ERROR: seed file could not be validated. See errors for details.\n")
			os.Exit(1)
		}
	}

	// Open and parse seed file into struct
	seedFile, err := os.Open(seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error opening %s. Error received is: %s\n", seedFileName, err.Error())
		os.Exit(1)
	}
	jsonParser := json.NewDecoder(seedFile)
	var seed objects.Seed
	if err = jsonParser.Decode(&seed); err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: A valid %s must be present in the working directory. Error parsing %s.\nError received is: %s\n",
			constants.SeedFileName, seedFileName, err.Error())
		os.Exit(2)
	}

	// Build Docker image
	if buildCmd.Parsed() {
		DockerBuild(&seed, "")
	}

	// Run Docker image
	if runCmd.Parsed() {
		DockerRun(&seed)
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
		fmt.Fprintf(os.Stderr, "ERROR: Error attaching to std output pipe. %s\n", err.Error())
	}

	dCmd.Start()
	if string(dErr.Bytes()) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n", string(dErr.Bytes()))
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
func DockerBuild(seed *objects.Seed, imageName string) {

	// Retrieve docker image name
	if imageName == "" {
		imageName = BuildImageName(seed)
	}

	// Set the seed.manifest.json contents as an image label
	label := "com.ngageoint.seed.manifest=" + GetManifestLabel()

	jobDirectory := buildCmd.Lookup(constants.JobDirectoryFlag).Value.String()

	// Build Docker image
	fmt.Fprintf(os.Stderr, "INFO: Building %s\n", imageName)
	buildCmd := exec.Command("docker", "build", "-t", imageName, jobDirectory, "--label", label)

	// attach stderr pipe
	errPipe, err := buildCmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error attaching to build command stderr. %s\n", err.Error())
	}

	// Attach stdout pipe
	outPipe, err := buildCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error attaching to build command stdout. %s\n", err.Error())
	}

	// Run docker build
	if err := buildCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error executing docker build. %s\n", err.Error())
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
		os.Exit(2)
	}
}

//GetManifestLabel returns the seed.manifest.json as LABEL com.ngageoint.seed.manifest contents
func GetManifestLabel() string {
	// read the seed.manifest.json into a string
	seedbytes, err := ioutil.ReadFile(seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Eror reading %s. %s\n", seedFileName, err.Error())
		os.Exit(2)
	}
	var seedbuff bytes.Buffer
	json.Compact(&seedbuff, seedbytes)
	seedbytes, err = json.Marshal(seedbuff.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error marshalling seed manifest. %s\n", err.Error())
	}

	// Escape forward slashes and dollar signes
	seed := string(seedbytes)
	seed = strings.Replace(seed, "$", "\\$", -1)
	seed = strings.Replace(seed, "/", "\\/", -1)

	return seed
}

//DockerRun Runs image described by Seed spec
func DockerRun(seed *objects.Seed) {

	// Builds the image name
	imageName := BuildImageName(seed)

	// Test if image has been built; Rebuild if not
	imgsArgs := []string{"images", "-q", imageName}
	imgOut, err := exec.Command("docker", imgsArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker %v\n", imgsArgs)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	} else if string(imgOut) == "" {
		fmt.Fprintf(os.Stderr, "INFO: No docker image found for image name %s. Building image now...\n", imageName)
		DockerBuild(seed, imageName)
	}

	// build docker run command
	dockerArgs := []string{"run"}

	if runCmd.Lookup(constants.RmFlag).Value.String() == "true" {
		dockerArgs = append(dockerArgs, "--rm")
	}

	var mountsArgs []string

	// expand INPUT_FILEs to specified inputData files
	if runCmd.Lookup(constants.InputDataFlag).Value.String() != "" {
		inMounts := DefineInputs(seed)
		if inMounts != nil {
			mountsArgs = append(mountsArgs, inMounts...)
		}
	}

	// mount the JOB_OUTPUT_DIR (outDir flag)
	var outDir string
	if runCmd.Lookup(constants.JobOutputDirFlag).Value.String() != "" {
		outDir = SetOutputDir(seed)
		if outDir != "" {
			mountsArgs = append(mountsArgs, "-v")
			mountsArgs = append(mountsArgs, outDir+":"+outDir)
		}
	}
	// Settings
	settings := DefineSettings(seed)
	_ = settings

	// Set any defined environment variables
	envVars := DefineEnvironmentVariables(seed)
	_ = envVars

	// Additional Mounts defined in seed.json
	mounts := DefineMounts(seed)
	_ = mounts

	// Build Docker command arguments:
	// 		run
	//		-rm if specified
	// 		all mounts
	//		image name
	//		Job.Interface.Args
	dockerArgs = append(dockerArgs, mountsArgs...)
	dockerArgs = append(dockerArgs, imageName)

	// Parse out command arguments from seed.Job.Interface.Args
	args := strings.Split(seed.Job.Interface.Args, " ")
	dockerArgs = append(dockerArgs, args...)

	// Run
	var cmd bytes.Buffer
	cmd.WriteString("docker ")
	for _, s := range dockerArgs {
		cmd.WriteString(s + " ")
	}
	fmt.Fprintf(os.Stderr, "\nINFO: Running Docker command:\n%s\n", cmd.String())

	// Run Docker command and capture output
	runCmd := exec.Command("docker", dockerArgs...)
	// attach stderr pipe
	errPipe, err := runCmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error attaching to run command stderr. %s\n", err.Error())
	}

	// Attach stdout pipe
	outPipe, err := runCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error attaching to run command stdout. %s\n", err.Error())
	}

	// Run docker build
	if err := runCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error executing docker run. %s\n", err.Error())
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		fmt.Fprintf(os.Stdout, "%s\n", string(slurp))
	}

	// check for errors on stderr
	slurperr, _ := ioutil.ReadAll(errPipe)
	if string(slurperr) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error running image '%s':\n%s\n",
			imageName, string(slurperr))
		os.Exit(2)
	}

	// Validate output against pattern
	if seed.Job.Interface.OutputData.Files != nil ||
		seed.Job.Interface.OutputData.JSON != nil {
		ValidateOutput(seed, outDir)
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
	validateCmd.StringVar(&schema, constants.SchemaFlag, "", "JSON schema file to validate seed against.")
	validateCmd.StringVar(&schema, constants.ShortSchemaFlag, "", "JSON schema file to validate seed against.")

	validateCmd.Usage = func() {
		PrintValidateUsage()
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
	default:
		fmt.Fprintf(os.Stderr, "%q is not a valid command.\n", os.Args[1])
		PrintUsage()
		os.Exit(2)
	}
}

//PrintCommandUsage prints usage of parsed command, or seed usage if no command parsed
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
	fmt.Fprintf(os.Stderr, "  validate\tValidates a Seed spec")
	fmt.Fprintf(os.Stderr, "\nRun 'seed COMMAND --help' for more information on a command.\n")
	os.Exit(1)
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintBuildUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\tseed build [-d JOB_DIRECTORY]\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory containing Seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	os.Exit(1)
}

//PrintRunUsage prints the seed run usage arguments, then exits the program
func PrintRunUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed run [-i INPUT_KEY=INPUT_FILE ...] -o JOB_OUTPUT_DIRECTORY [OPTIONS]\n")
	fmt.Fprintf(os.Stderr, "\nRuns Docker image defined by seed spec.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory containing seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s\tSpecifies the key/value input data values of the seed spec in the format INPUT_FILE_KEY=INPUT_FILE_VALUE\n",
		constants.ShortInputDataFlag, constants.InputDataFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tDocker image name to run\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s   \tJob Output Directory Location\n",
		constants.ShortJobOutputDirFlag, constants.JobOutputDirFlag)
	fmt.Fprintf(os.Stderr, "  -%s            \tAutomatically remove the container when it exits (docker run --rm)\n",
		constants.RmFlag)
	os.Exit(1)
}

//PrintListUsage prints the seed list usage information, then exits the program
func PrintListUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed list [OPTIONS]\n")
	fmt.Fprintf(os.Stderr, "\nAllows for listing all Seed compliant images residing on the local system.\n")
	fmt.Fprintf(os.Stderr, "\nLists all '-seed' docker images on the local machine.\n")
	os.Exit(1)
}

//PrintSearchUsage prints the seed search usage information, then exits the program
func PrintSearchUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed search [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nAllows for discovery of seed compliant images hosted within a Docker registry.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -r -repo Specifies a specific registry to search (default is docker.io)\n")
	os.Exit(1)
}

//PrintPublishUsage prints the seed publish usage information, then exits the program
func PrintPublishUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed publish [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nAllows for discovery of seed compliant images hosted within a Docker registry.\n")
	// fmt.Fprintf(os.Stderr, "\nOptions:\n")
	// fmt.Fprintf(os.Stderr, "  -r, -repo Specifies a specific registry to search (default is docker.io)\n")
	os.Exit(1)
}

//PrintValidateUsage prints the seed validate usage, then exits the program
func PrintValidateUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed validate [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nValidates the given %s is compliant with the Seed spec.\n", constants.SeedFileName)
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies directory in which Seed is located (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s   \tExternal Seed schema file; Overrides built in schema to validate Seed spec against\n",
		constants.ShortSchemaFlag, constants.SchemaFlag)
	os.Exit(1)
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
func DefineInputs(seed *objects.Seed) []string {

	// TODO: validate number of inputData flags to number of Interface.InputData.Files
	var mountArgs []string
	inputStr := runCmd.Lookup(constants.InputDataFlag).Value.String()
	var inputs []string
	inputs = strings.Split(inputStr, ",")

	for _, f := range inputs {
		x := strings.Split(f, "=")
		if len(x) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: Input files should be specified in KEY=VALUE format.\n")
			fmt.Fprintf(os.Stderr, "ERROR: Unknown key for input %v encountered.\n", inputs)
			continue
		}

		key := x[0]
		val := x[1]

		// Expand input VALUE
		val = GetFullPath(val)

		// Replace key if found in args strings
		// Handle replacing KEY or ${KEY} or $KEY
		seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args, "${"+key+"}", val, -1)
		seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args, "$"+key, val, -1)
		seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args, key, val, -1)

		for _, k := range seed.Job.Interface.InputData.Files {
			if k.Name == key {
				mountArgs = append(mountArgs, "-v")
				mountArgs = append(mountArgs, val+":"+val)
			}
		}
	}

	return mountArgs
}

//SetOutputDir replaces the JOB_OUTPUT_DIR argument with the given output directory.
// Returns output directory string
func SetOutputDir(seed *objects.Seed) string {
	outputDir := runCmd.Lookup(constants.JobOutputDirFlag).Value.String()
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

	seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args,
		"$JOB_OUTPUT_DIR", outdir, -1)
	seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args,
		"${JOB_OUTPUT_DIR}", outdir, -1)
	return outdir
}

//DefineEnvironmentVariables defines any seed specified environment variables.
func DefineEnvironmentVariables(seed *objects.Seed) []string {
	if seed.Job.Interface.EnvVars != nil {
		var envVars []string
		for _, envVar := range seed.Job.Interface.EnvVars {
			envVars = append(envVars, "-e")
			envVars = append(envVars, envVar.Name)
			envVars = append(envVars, envVar.Value)
		}
		return envVars
	}
	return nil
}

//DefineMounts defines any seed specified mounts. TODO
func DefineMounts(seed *objects.Seed) []string {
	if seed.Job.Interface.Mounts != nil {
		/*
			fmt.Println("Found some mounts....")
			var mounts []string
			for _, mount := range seed.Job.Interface.Mounts {
				fmt.Println(mount.Name)
				mounts = append(mounts, "-v")
				mountPath := mount.Path + ":" + mount.Path

				if mount.Mode != "" {
					mountPath += ":" + mount.Mode
				}
				dockerArgs = append(dockerArgs,mountPath)
			}
		*/
		return nil
	}

	return nil
}

//DefineSettings defines any seed specified docker settings. TODO
// Return []string of docker command arguments in form of:
//	"-?? setting1=val1 -?? setting2=val2 etc"
func DefineSettings(seed *objects.Seed) []string {
	return nil
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

			// find all pattern matches in JOB_OUTPUT_DIR
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
	// Look for ResultsFileManifestName.json in the root of the JOB_OUTPUT_DIR
	// and then validate any keys identified in OutputData exist
	if seed.Job.Interface.OutputData.JSON != nil {

		fmt.Fprintf(os.Stderr, "INFO: Validating %s...\n",
			filepath.Join(outDir, constants.ResultsFileManifestName))
		// look for results manifest
		outDir := runCmd.Lookup(constants.JobOutputDirFlag).Value.String()
		manfile := filepath.Join(outDir, constants.ResultsFileManifestName)
		if _, err := os.Stat(manfile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %s specified but cannot be found. %s\n Exiting testrunner.\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}

		bites, err := ioutil.ReadFile(filepath.Join(outDir, constants.ResultsFileManifestName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error reading %s.%s\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}

		// Read in manifest
		var mfile interface{}
		json.Unmarshal(bites, &mfile)
		mf := mfile.(map[string]interface{})

		// Loop through defined name/key values to extract from results_manifest.json
		for _, jsonStr := range seed.Job.Interface.OutputData.JSON {
			// name
			name := jsonStr.Name
			// key
			key := jsonStr.Key

			if key != "" {
				if mf[key] != nil {
					fmt.Fprintf(os.Stderr, "SUCCESS: Key/Value found: %s=%v\n", key, mf[key])
					// error if not found
				} else {
					fmt.Fprintf(os.Stderr, "ERROR: No output value found for key %s in results_manifest.json\n", key)
				}
			} else {
				if mf[name] != nil {
					fmt.Fprintf(os.Stderr, "SUCCESS: Name/Value found: %s=%v\n", name, mf[name])

					// error if not found
				} else {
					fmt.Fprintf(os.Stderr, "ERROR: No value found for name %s\n", name)
				}
			}
		}
	}
}

//ValidateSeedFile Validates the seed.manifest.json file based on the given schema
func ValidateSeedFile(schemaFile string, seedFileName string) bool {
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
		fmt.Fprintf(os.Stderr, "INFO: Validating seed file %s against schema...\n", seedFileName)
		schemaBytes, _ := constants.Asset("data/seed.manifest.schema.json")
		schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
		docLoader := gojsonschema.NewReferenceLoader("file://" + seedFileName)
		result, err = gojsonschema.Validate(schemaLoader, docLoader)
	}

	// Error occurred loading the schema or seed.manifest.json
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR: Error validating seed file against schema. Error is: %s\n", err.Error())
		return false
	}

	// Validation failed. Print results
	if !result.Valid() {
		fmt.Fprintf(os.Stderr, "\nERROR: %s is not valid. See errors:\n", seedFileName)
		for _, e := range result.Errors() {
			fmt.Fprintf(os.Stderr, "-ERROR %v\n", e.Description())
			fmt.Fprintf(os.Stderr, "\tField: %s\n", e.Field())
			fmt.Fprintf(os.Stderr, "\tContext: %s\n", e.Context().String())
		}
		fmt.Fprintf(os.Stderr, "\n")
		return false
	}

	// Validation succeeded
	fmt.Fprintf(os.Stderr, "\nSUCCESS: %s is valid.\n\n", seedFileName)
	return true
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
