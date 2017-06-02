package objects

type SeedJob_1_4 struct {
	Name string
	Version string
	AlgorithmVersion string
	PackageVersion string
	Title string
	Description string
	Tag []string
	AuthorName string
	AuthorEmail string
	AuthorUrl string
	Timeout int
	Cpus float64
	Mem float64
	SharedMem float64
	Storage float64
	Interface JobInterface
	ErrorMapping ErrorMap
}

type JobInterface struct {
	Cmd string
	Args string
	InputData InputData
	OutputData OutputData
	EnvVars []JobEnvVar
	Mounts []JobMount
	Settings []JobSetting
}

type InputData struct {
	Files []InFile
	Json []JobJson
}

type InFile struct {
	Name string
	Required bool
	MediaType []string
}

type OutputData struct {
	Files []OutFile
	Json []JobJson
}

type OutFile struct {
	Name string
	MediaType string
	Count string
	Pattern string
}

type JobJson struct {
	Name string
	Key string
	Type string
}

type JobEnvVar struct {
	Name string
	Value string
}

type JobMount struct {
	Name string
	Path string
	Mode string
}

type JobSetting struct {
	Name string
	Secret bool
}

type ErrorMap struct {
	Code int
	Title string
	Description string
	Category string
}
