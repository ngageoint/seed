package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	file := GetArgs()
	
	jsonText, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
	}
	var b bytes.Buffer
	err = json.Compact(&b, jsonText)
	if err != nil {
		fmt.Println(err)
	}
	slashes := bytes.Replace(b.Bytes(), []byte("/"), []byte("\\/"), -1)
	quotes := bytes.Replace(slashes, []byte("\""), []byte("\\\""), -1)
	dollars := bytes.Replace(quotes, []byte("$"), []byte("\\$"), -1)
	fmt.Println(string(dollars))
}

// Retrieve command line arguments
func GetArgs() string {
	var file string

	flag.StringVar(&file, "file", "", "Json file to shrink")

	flag.Parse()
	if (len(file) == 0) {
		fmt.Println("\n\"JsonCompactor\" requires a file to be specified \n")
		os.Exit(1)
	}
	return file
}

