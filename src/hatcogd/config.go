// Configuration file parser
package main

import (
	"io/ioutil"
	"log"
	"strings"
)

type Config map[string]string

func LoadConfig(filename string) (Config, error) {

	var parts []string
	var key, value string

	config := make(map[string]string)

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") { // A comment
			continue
		}
		parts = strings.SplitN(line, "=", 2)
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
		config[key] = strings.Trim(value, "\"'")
	}

	return config, nil
}

func (self Config) Get(key string) string {
	val := self[key]
	if len(val) == 0 {
		log.Fatal("Missing configuration for '" + key + "'")
	}
	return val
}
