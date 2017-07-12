/*
Package seedrunner implements a command line interface  library to build and run
docker images defined by a seed.manifest.json file.
usage is as folllows:
	seed build [OPTIONS]
		options:
		-d, -directory	The directory containing the seed spec and Dockerfile
										(default is current directory)

	seed run [OPTION]
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
*/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var buildCmd *flag.FlagSet
var runCmd *flag.FlagSet
var directory string

//JobNameFlag Defines the job name flag
// const JobNameFlag = "jobName"

//ShortJobNameFlag Defines the shorthand job name flag
// const ShortJobNameFlag = "j"

//AlgVersionFlag defines the algorithm version
// const AlgVersionFlag = "algVersion"

//ShortAlgVersionFlag defines the shorthand algorithm version flag
// const ShortAlgVersionFlag = "a"

//PkgVersionFlag defines the PkgVersionFlag
// const PkgVersionFlag = "pkgVersion"

//ShortPkgVersionFlag defines the shorthand package version flag
// const ShortPkgVersionFlag = "p"

//JobDirectoryFlag defines the location of the seed spec and Dockerfile
const JobDirectoryFlag = "directory"

//ShortJobDirectoryFlag defines the shorthand location of the seed spec and Dockerfile
const ShortJobDirectoryFlag = "d"

//InputDataFlag defines the InputDataFlag
const InputDataFlag = "inputData"

//ShortInputDataFlag defines the shorthand input data flag
const ShortInputDataFlag = "i"

//JobOutputDirFlag defines the job output directory
const JobOutputDirFlag = "outDir"

//ShortJobOutputDirFlag defines the shorthand output directory
const ShortJobOutputDirFlag = "o"

//ShortImgNameFlag defines image name to run
const ShortImgNameFlag = "in"

//ImgNameFlag defines image name to run
const ImgNameFlag = "imageName"

//RmFlag defines if the docker image should be removed after docker run is executed
const RmFlag = "rm"

//SeedFileName defines the filename for the seed file
const SeedFileName = "seed.manifest.json"

//ResultsFileManifestName defines the filename for the results_manifest file
const ResultsFileManifestName = "results_manifest.json"

type arrayFlags []string

//String converts an arrayFlags object to a single, comma separated string.
func (flags *arrayFlags) String() string {
	var buff bytes.Buffer
	for i, f := range *flags {
		buff.WriteString(f)

		if i < (len(*flags) - 1) {
			buff.WriteString(",")
		}
	}
	return buff.String()
}

//Set defines the setter function for *arrayFlags.
func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

//Seed represents a seed.manifest.json object.
type Seed struct {
	ManifestVersion string `json:"manifest_version"`
	Job             struct {
		Name             string `json:"name"`
		AlgorithmVersion string `json:"algorithmVersion"`
		PackageVersion   string `json:"packageVersion"`
		Interface        struct {
			Args      string `json:"args"`
			InputData struct {
				Files []struct {
					Name      string   `json:"name"`
					MediaType []string `json:"mediaType"`
					Pattern   string   `json:"pattern"`
					Path      string   `json:"path"`
				}
			}
			OutputData struct {
				Files []struct {
					Name      string   `json:"name"`
					MediaType []string `json:"mediaType"`
					Pattern   string   `json:"pattern"`
					Count     string   `json:"count"`
					Required  bool     `json:"required"`
				}
				Json []struct {
					Name     string `json:"name"`
					Type     string `json:"type"`
					Key      string `json:"key"`
					Required bool   `json:"required"`
				}
			}
			EnvVars []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			}
			Mounts []struct {
				Name string `json:"name"`
				Path string `json:"path"`
				Mode string `json:"mode"`
			}
			Settings []struct {
				Name   string `json:"name"`
				Secret string `json:"secret"`
			}
			ErrorMapping []struct {
				Code        int    `json:"code"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Category    string `json:"category"`
			}
		}
	}
}

/* Run command defaults:
   image name and tag: if not specified, attempt to guess from CWD if a
      seed.json exists. Otherwise error.
   input: no default; args should match file InputData as described in
      seed.json. You'll need to search / replace this with container
      resolvable paths. It's the algorithm developers respobsibility to
      create parameter expansion
   output: no default; single directory where output files are placed. Glob
      capture expressions are described in seed.json
*/
func main() {

	// Parse input flags
	DefineFlags()

	seedFileName := SeedFileName
	if directory != "." {
		seedFileName = path.Join(directory, SeedFileName)
	}

	// Verify seed.json exists If not, error
	if _, err := os.Stat(seedFileName); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: %s cannot be found. Exiting seed testrunner...\n", seedFileName)
		os.Exit(1)
	}

	// Parse through seed file
	seedFile, err := os.Open(seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err.Error())
		os.Exit(1)
	}
	jsonParser := json.NewDecoder(seedFile)
	var seed Seed
	if err = jsonParser.Decode(&seed); err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: A valid %s must be present in the working directory. Error parsing %s",
			SeedFileName, SeedFileName)
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(2)
	}

	// Build Docker image
	if buildCmd.Parsed() {
		DockerBuild(&seed, "")
	}

	if runCmd.Parsed() {
		DockerRun(&seed)
	}
}

//DockerBuild Builds the docker image with the given image tag.
func DockerBuild(seed *Seed, imageName string) {

	// Retrieve docker image name
	if imageName == "" {
		imageName = BuildImageName(seed)
	}

	jobDirectory := buildCmd.Lookup(JobDirectoryFlag).Value.String()

	// Build Docker image
	println("Building " + imageName)
	buildOutput, err := exec.Command("docker", "build", "-t", imageName, jobDirectory).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err.Error())
	} else {
		fmt.Fprintf(os.Stdout, string(buildOutput))
	}
}

//DockerRun Runs the provided docker command.
func DockerRun(seed *Seed) {

	// Builds the image name
	imageName := BuildImageName(seed)

	// Test if image has been built
	imgsArgs := []string{"images", "-q", imageName}
	imgOut, err := exec.Command("docker", imgsArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing docker %v\n", imgsArgs)
		fmt.Fprintf(os.Stderr, err.Error())
	} else if string(imgOut) == "" {
		fmt.Fprintf(os.Stderr, "INFO: No docker image found for image name %s. Building image now...\n", imageName)
		DockerBuild(seed, imageName)
	}

	// build docker run command
	dockerArgs := []string{"run"}

	if runCmd.Lookup(RmFlag).Value.String() == "true" {
		dockerArgs = append(dockerArgs, "--rm")
	} //, imageName}
	var mountsArgs []string

	// expand INPUT_FILEs to specified inputData files
	if runCmd.Lookup(InputDataFlag).Value.String() != "" {
		inMounts := DefineInputs(seed)
		if inMounts != nil {
			mountsArgs = append(mountsArgs, inMounts...)
		}
	}

	// mount the JOB_OUTPUT_DIR (outDir flag)
	if runCmd.Lookup(JobOutputDirFlag).Value.String() != "" {
		outDir := SetOutputDir(seed)
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
	fmt.Fprintf(os.Stderr, "\nINFO: Running Docker command:\n %s\n\n", cmd.String())

	// Run Docker command and capture output
	runOutput, err := exec.Command("docker", dockerArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error running docker command.\n%s\n", err.Error())
	} else {
		fmt.Fprintf(os.Stdout, "%s", string(runOutput))
	}

	// Validate output against pattern
	if seed.Job.Interface.OutputData.Files != nil ||
		seed.Job.Interface.OutputData.Json != nil {
		ValidateOutput(seed)
	}
}

//DefineFlags defines the flags available for the seed runner.
// Possible flags include:
// seed build optional arguments:
//		-j, -jobName			The job name
//		-a, -algVersion		The algorithm version
//		-p, -pkgVersion		The package version
// seed run required arguments:
// 		-i, -inputData		One or more input data files
//		-o, -outDir				Output directory
//		-rm								Removes the image after running
func DefineFlags() {

	// build command flags
	buildCmd = flag.NewFlagSet("build", flag.ContinueOnError)
	buildCmd.StringVar(&directory, JobDirectoryFlag, ".", "Directory of seed spec and Dockerfile (default is current directory).")
	buildCmd.StringVar(&directory, ShortJobDirectoryFlag, ".", "Directory of seed spec and Dockerfile (default is current directory).")

	// Print usage function
	buildCmd.Usage = func() {
		PrintBuildUsage()
	}

	// Run command flags
	runCmd = flag.NewFlagSet("run", flag.ContinueOnError)
	runCmd.StringVar(&directory, JobDirectoryFlag, ".", "Location of the seed spec and Dockerfile")
	runCmd.StringVar(&directory, ShortJobDirectoryFlag, ".", "Location of the seed spec and Dockerfile")

	var imgNameFlag string
	runCmd.StringVar(&imgNameFlag, ImgNameFlag, "", "Name of Docker image to run")
	runCmd.StringVar(&imgNameFlag, ShortImgNameFlag, "", "Name of Docker image to run")

	var inputs arrayFlags
	runCmd.Var(&inputs, InputDataFlag, "Defines any input data arguments")
	runCmd.Var(&inputs, ShortInputDataFlag, "Defines input data arguments")

	var outdir string
	runCmd.StringVar(&outdir, JobOutputDirFlag, "$PWD", "outDir ${JOB_OUTPUT_DIR}")
	runCmd.StringVar(&outdir, ShortJobOutputDirFlag, "$PWD", "outDir ${JOB_OUTPUT_DIR}")

	var rmVar bool
	runCmd.BoolVar(&rmVar, RmFlag, false, "Specifying the -rm flag automatically removes the image after executing docker run")

	// Run usage function
	runCmd.Usage = func() {
		PrintRunUsage()
	}

	// List command
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.Usage = func() {
		PrintListUsage()
	}

	// Search command
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchCmd.Usage = func() {
		PrintSearchUsage()
	}

	if len(os.Args) == 1 {
		PrintUsage()
	}

	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		if len(os.Args) > 2 && os.Args[2] != "-d" {
			directory = os.Args[2]
		}
	case "run":
		runCmd.Parse(os.Args[2:])
	case "search":
		fmt.Fprintf(os.Stderr, "%q is not yet implemented\n\n", os.Args[1])
		PrintSearchUsage()
	case "list":
		fmt.Fprintf(os.Stderr, "%q is not yet implemented\n\n", os.Args[1])
		PrintListUsage()
	default:
		fmt.Fprintf(os.Stderr, "%q is not a valid command.\n", os.Args[1])
		os.Exit(2)
	}
}

//PrintUsage prints the seed usage arguments
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\tseed COMMAND\n\n")
	fmt.Fprintf(os.Stderr, "A test runner for seed spec compliant algorithms\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  build \tBuilds a docker image based on the seed spec\n")
	fmt.Fprintf(os.Stderr, "  list  \tLists all seed images found on local Docker daemon\n")
	fmt.Fprintf(os.Stderr, "  run   \tRuns docker image based on the seed spec. Also builds docker image if not found\n")
	fmt.Fprintf(os.Stderr, "  search\tSearches the docker registry for -seed images (default docker.io)\n")
	fmt.Fprintf(os.Stderr, "\nRun 'seed COMMAND --help' for more information on a command.\n")
	os.Exit(1)
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintBuildUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\tseed build [-d JOB_DIRECTORY]\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s,  -%s\tDirectory containing seed spec and Dockerfile (default is current directory)\n",
		ShortJobDirectoryFlag, JobDirectoryFlag)
	os.Exit(1)
}

//PrintRunUsage prints the seed run usage arguments, then exits the program
func PrintRunUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed run [-i INPUT_KEY=INPUT_FILE ...] -o JOB_OUTPUT_DIRECTORY [OPTIONS]\n")
	fmt.Fprintf(os.Stderr, "\nRuns Docker image. Input data defined in seed spec must be specified via the -i option.\n\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s,  -%s\tDirectory containing seed spec and Dockerfile (default is current directory)\n",
		ShortJobDirectoryFlag, JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s,  -%s\tSpecifies the key/value input data values of the seed spec in the format INPUT_FILE_KEY=INPUT_FILE_VALUE\n", ShortInputDataFlag, InputDataFlag)
	fmt.Fprintf(os.Stderr, "  -%s, -%s\tDocker image name to run\n", ShortImgNameFlag, ImgNameFlag)
	fmt.Fprintf(os.Stderr, "  -%s,  -%s   \tJob Output Directory Location\n",
		ShortJobOutputDirFlag, JobOutputDirFlag)
	fmt.Fprintf(os.Stderr, "  -%s            \tAutomatically remove the container when it exits (docker run --rm)\n", RmFlag)
	os.Exit(1)
}

//PrintListUsage prints the seed list usage information, then exits the program
func PrintListUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed list \n")
	fmt.Fprintf(os.Stderr, "\nLists all -seed docker images on the local machine.\n")
	os.Exit(1)
}

//PrintSearchUsage prints the seed search usage information, then exits the program
func PrintSearchUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed search [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nSearches the remote registry for -seed docker images.\n")
	fmt.Fprintf(os.Stderr, "  -r, -repo Specifies a specific registry to search (default is docker.io)\n")
	os.Exit(1)
}

//BuildImageName extracts the Docker Image name from either the input arguments
// (via commnad flags -jobName -algVersion -pkgVersion) or the seed.json, or
// combination of th etwo. Returns image name in the form of
// 	jobName-algVersion-seed:pkgVersion
func BuildImageName(seed *Seed) string {
	var buffer bytes.Buffer

	if runCmd.Parsed() && runCmd.Lookup(ImgNameFlag).Value.String() != "" {
		buffer.WriteString(runCmd.Lookup(ImgNameFlag).Value.String())
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
func DefineInputs(seed *Seed) []string {

	// TODO: validate number of inputData flags to number of Interface.InputData.Files
	var mountArgs []string
	inputStr := runCmd.Lookup(InputDataFlag).Value.String()
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

		// Replace key if found in args strings
		// Handle replacing KEY or ${KEY} or $KEY
		seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args, "${"+key+"}", val, -1)
		seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args, "$"+key, val, -1)
		seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args, key, val, -1)

		for _, k := range seed.Job.Interface.InputData.Files {
			if k.Name == key {
				k.Path = val
				mountArgs = append(mountArgs, "-v")
				mountArgs = append(mountArgs, k.Path+":"+k.Path)
			}
		}
	}

	return mountArgs
}

//SetOutputDir replaces the JOB_OUTPUT_DIR argument with the given output directory.
// Returns output directory string
func SetOutputDir(seed *Seed) string {
	outputDir := runCmd.Lookup(JobOutputDirFlag).Value.String()
	seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args,
		"$JOB_OUTPUT_DIR", outputDir, -1)
	seed.Job.Interface.Args = strings.Replace(seed.Job.Interface.Args,
		"${JOB_OUTPUT_DIR}", outputDir, -1)
	return outputDir
}

//DefineEnvironmentVariables defines any seed specified environment variables.
func DefineEnvironmentVariables(seed *Seed) []string {
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
func DefineMounts(seed *Seed) []string {
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
func DefineSettings(seed *Seed) []string {
	return nil
}

///ValidateOutput validates the output of the docker run command. Output data is
// validated as defined in the seed.Job.Interface.OutputData.
func ValidateOutput(seed *Seed) {
	// Validate any OutputData.Files
	if seed.Job.Interface.OutputData.Files != nil {
		outDir := runCmd.Lookup(JobOutputDirFlag).Value.String()

		// For each defined OutputData file:
		//	#1 Check file media type
		// 	#2 Check file names match output pattern
		//  #3 Check number of files (if defined)
		for _, f := range seed.Job.Interface.OutputData.Files {

			// find all pattern matches in JOB_OUTPUT_DIR
			matches, _ := filepath.Glob(path.Join(outDir, f.Pattern))

			// Check media type of matches
			//fmt.Println("Found " + strconv.Itoa(len(matches)) + " matches for pattern '" + f.Pattern + "':")
			count := 0
			var matchList []string
			for _, match := range matches {
				ext := path.Ext(match)
				mType := mime.TypeByExtension(ext)
				if contains(f.MediaType, mType) {
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
	if seed.Job.Interface.OutputData.Json != nil {
		// look for results manifest
		outDir := runCmd.Lookup(JobOutputDirFlag).Value.String()
		manfile := path.Join(outDir, ResultsFileManifestName)
		if _, err := os.Stat(manfile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %s specified but cannot be found. Exiting testrunner...\n", ResultsFileManifestName)
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}

		bites, err := ioutil.ReadFile(path.Join(outDir, ResultsFileManifestName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error reading %s\n", ResultsFileManifestName)
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}

		// Read in manifest
		var mfile interface{}
		json.Unmarshal(bites, &mfile)
		mf := mfile.(map[string]interface{})

		// Loop through defined name/key values to extract from results_manifest.json
		for _, jsonStr := range seed.Job.Interface.OutputData.Json {
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

func contains(array []string, str string) bool {
	for _, a := range array {
		if strings.Contains(str, a) || strings.Contains(a, str) {
			return true
		}
	}
	return false
}
