package objects

import (
  "bytes"
)

type ArrayFlags []string

//String converts an arrayFlags object to a single, comma separated string.
func (flags *ArrayFlags) String() string {
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
func (flags *ArrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}
