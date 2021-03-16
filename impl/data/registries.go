package data

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var registryMap map[string]map[string]interface{}

func RegistryID(registry string, entry string) int32 {
	if registryMap == nil {
		registryMap = make(map[string]map[string]interface{})
		// Open our jsonFile
		jsonFile, _ := os.Open("./data-generator/generated/reports/registries.json")
		byteValue, _ := ioutil.ReadAll(jsonFile)

		_ = json.Unmarshal(byteValue, &registryMap)
		jsonFile.Close()
	}

	entries := registryMap[registry]["entries"].(map[string]interface{})
	return int32(entries[entry].(map[string]interface{})["protocol_id"].(float64))
}
