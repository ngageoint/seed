package objects

import (
	"encoding/json"
)

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
	ErrorMapping []ErrorMap
}

func (o *SeedJob_1_4) UnmarshalJSON(b []byte) error {
	type xjob SeedJob_1_4
	xo := &xjob{SharedMem: 0.0, Storage: 0.0}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = SeedJob_1_4(*xo)
	return nil
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
	Json []InJson
}

type InFile struct {
	Name string
	Multiple bool
	Required bool
	MediaType []string
}

func (o *InFile) UnmarshalJSON(b []byte) error {
	type xInFile InFile
	xo := &xInFile{Multiple: false, Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = InFile(*xo)
	return nil
}

type InJson struct {
	Name string
	Type string
	Required bool
}

func (o *InJson) UnmarshalJSON(b []byte) error {
	type xInJson InJson
	xo := &xInJson{Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = InJson(*xo)
	return nil
}

type OutputData struct {
	Files []OutFile
	Json []OutJson
}

type OutFile struct {
	Name string
	MediaType string
	Count string
	Pattern string
	Required bool
}

func (o *OutFile) UnmarshalJSON(b []byte) error {
	type xOutFile OutFile
	xo := &xOutFile{Count: "1", Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = OutFile(*xo)
	return nil
}

type OutJson struct {
	Name string
	Key string
	Type string
	Required bool
}

func (o *OutJson) UnmarshalJSON(b []byte) error {
	type xOutJson OutJson
	xo := &xOutJson{Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = OutJson(*xo)
	return nil
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

func (o *JobMount) UnmarshalJSON(b []byte) error {
	type xJobMount JobMount
	xo := &xJobMount{Mode: "ro"}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = JobMount(*xo)
	return nil
}

type JobSetting struct {
	Name string
	Secret bool
}

func (o *JobSetting) UnmarshalJSON(b []byte) error {
	type xJobSetting JobSetting
	xo := &xJobSetting{Secret: false}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = JobSetting(*xo)
	return nil
}

type ErrorMap struct {
	Code int
	Title string
	Description string
	Category string
}

func (o *ErrorMap) UnmarshalJSON(b []byte) error {
	type xErrorMap ErrorMap
	xo := &xErrorMap{Category: "algorithm"}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = ErrorMap(*xo)
	return nil
}
