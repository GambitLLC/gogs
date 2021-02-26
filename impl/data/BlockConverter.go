package data

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var blocksMap map[string]map[string]interface{}
var idMap = make(map[string]map[string]int64)

func ParseBlockId(name string, properties map[string]interface{}) int64 {
	if blocksMap == nil {
		blocksMap = make(map[string]map[string]interface{})
		// Open our jsonFile
		jsonFile, _ := os.Open("./data-generator/generated/reports/blocks.json")
		byteValue, _ := ioutil.ReadAll(jsonFile)

		_ = json.Unmarshal(byteValue, &blocksMap)
		jsonFile.Close()
	}

	if idMap[name] == nil {
		idMap[name] = make(map[string]int64)
	}

	obj, _ := json.Marshal(properties)
	val, exists := idMap[name][string(obj)]
	if !exists {
		block := blocksMap[name]
		index := int64(0)
		for k, v := range properties {
			index *= 2
			propertyVals := block["properties"].(map[string]interface{})[k].([]interface{})
			for i, val := range propertyVals {
				if v == val {
					index += int64(i)
					break
				}
			}
		}

		id := int64(block["states"].([]interface{})[index].(map[string]interface{})["id"].(float64))
		idMap[name][string(obj)] = id
		return id
	}

	return val
}
