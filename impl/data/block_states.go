package data

import (
	"encoding/json"
	"gogs/impl/logger"
	"io/ioutil"
	"os"
)

var blocksMap map[string]map[string]interface{}
var idMap = make(map[string]map[string]int32)

func BlockStateID(name string, properties map[string]interface{}) int32 {
	if blocksMap == nil {
		blocksMap = make(map[string]map[string]interface{})
		// Open our jsonFile
		jsonFile, _ := os.Open("./data-generator/generated/reports/blocks.json")
		byteValue, _ := ioutil.ReadAll(jsonFile)

		_ = json.Unmarshal(byteValue, &blocksMap)
		jsonFile.Close()
	}

	if idMap[name] == nil {
		idMap[name] = make(map[string]int32)
	}

	obj, _ := json.Marshal(properties)
	val, exists := idMap[name][string(obj)]
	if !exists {
		block, exists := blocksMap[name]
		if !exists {
			logger.Printf("block state doesn't exist: %v", name)
			return 0
		}
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

		id := int32(block["states"].([]interface{})[index].(map[string]interface{})["id"].(float64))
		idMap[name][string(obj)] = id
		return id
	}

	return val
}
