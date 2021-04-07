package data

import (
	"encoding/json"
	"gogs/logger"
	"io/ioutil"
	"os"
)

var idMap map[string]map[string]int32

func BlockStateID(name string, properties map[string]interface{}) int32 {
	if idMap == nil {
		idMap = make(map[string]map[string]int32)
		blocksJson := make(map[string]map[string]interface{})

		jsonFile, _ := os.Open("./data-generator/generated/reports/blocks.json")
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)

		err := json.Unmarshal(byteValue, &blocksJson)
		if err != nil {
			return 0
		}

		// load all blocks into idMap
		for name, blockMap := range blocksJson {
			if idMap[name] == nil {
				idMap[name] = make(map[string]int32)
			}

			for _, state := range blockMap["states"].([]interface{}) {
				stateObj := state.(map[string]interface{})
				properties, exists := stateObj["properties"].(map[string]interface{})
				id := int32(stateObj["id"].(float64))
				if exists {
					obj, err := json.Marshal(properties)
					if err != nil {
						continue
					}
					idMap[name][string(obj)] = id
				}

				if isDefault, exists := stateObj["default"].(bool); exists && isDefault {
					idMap[name]["default"] = id
				}
			}
		}
	}

	if idMap[name] == nil {
		logger.Printf("block state doesn't exist: %s", name)
		return 0
	}

	if properties == nil {
		return idMap[name]["default"]
	}

	obj, _ := json.Marshal(properties)
	val, exists := idMap[name][string(obj)]
	if !exists {
		logger.Printf("block state %s with properties %v doesn't exist", name, properties)
		return 0
	}

	return val
}
