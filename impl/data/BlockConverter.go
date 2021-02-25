package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type jsonInterface map[string]interface{}
type ArrayStruct map[string][]struct {
	jsonInterface
}

var idMap = make(map[string]*int64)

func ParseBlockId(name string) int64 {
	if idMap[name] == nil {
		// Open our jsonFile
		jsonFile, err := os.Open("./data-generator/generated/reports/blocks.json")
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}

		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var result map[string]map[string][]map[string]interface{}
		json.Unmarshal([]byte(byteValue), &result)

		id := int64(result[name]["states"][len(result[name]["states"])-1]["id"].(float64))
		idMap[name] = &id
		return id
	}

	return *idMap[name]
}
