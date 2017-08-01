package objects

import (
	"encoding/json"
)

//Seed represents a seed.manifest.json object.
type Seed struct {
	SeedVersion string `json:"seedVersion"`
	Job         Job
}
	
type Job struct {
	Name             string `json:"name"`
	AlgorithmVersion string `json:"algorithmVersion"`
	PackageVersion   string `json:"packageVersion"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	AuthorName       string `json:"authorName"`
	AuthorEmail      string `json:"authorEmail"`
	AuthorUrl        string `json:"authorUrl"`
	Timeout          int    `json:"timeout"`
	Interface        Interface
	ErrorMapping     []ErrorMap
}

type Interface struct {
	Cmd        string `json:"cmd"`
	Resources  Resources
	InputData  InputData
	OutputData OutputData
	Mounts     []Mount
	Settings   []Setting
}

type Resources struct {
	Scalar []Scalar
}

type Scalar struct {
	Name            string  `json:"name"`
	Value           float64 `json:"value"`
	InputMultiplier float64 `json:"inputMultiplier"`
}

type InputData struct {
	Files []InFile
	Json  []InJson
}

type InFile struct {
	Name      string   `json:"name"`
	MediaType []string `json:"mediaType"`
	Multiple  bool     `json:"multiple"`
	Required  bool     `json:"required"`
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
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
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
	JSON  []OutJson
}

type OutFile struct {
	Name      string `json:"name"`
	MediaType string `json:"mediaType"`
	Count     string `json:"count"`
	Pattern   string `json:"pattern"`
	Required  bool   `json:"required"`
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
	Name     string `json:"name"`
	Key      string `json:"key"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
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

type Mount struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

func (o *Mount) UnmarshalJSON(b []byte) error {
	type xMount Mount
	xo := &xMount{Mode: "ro"}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = Mount(*xo)
	return nil
}

type Setting struct {
	Name   string `json:"name"`
	Secret bool   `json:"secret"`
}

func (o *Setting) UnmarshalJSON(b []byte) error {
	type xSetting Setting
	xo := &xSetting{Secret: false}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = Setting(*xo)
	return nil
}

type ErrorMap struct {
	Code        int    `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
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
