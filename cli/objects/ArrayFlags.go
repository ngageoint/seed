package objects

import (
	"bytes"
)

//ArrayFlags defines the values of a flag that may be used multiple times
//  i.e. -i KEY=VAL -i KEY2=VAL2 -i KEY3=VAL3 etc
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
