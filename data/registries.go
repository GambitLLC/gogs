package data

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var (
	protocolIDMap   = make(map[string]map[string]int32) // map[registry][namespacedID] = protocolID
	namespacedIDMap = make(map[string]map[int32]string) // map[registry][protocolID] = namespacedID
)

func ProtocolID(registry string, namespacedID string) int32 {
	if protocolIDMap[registry] == nil {
		if err := readRegistry(registry); err != nil {
			panic(err) // failed to open registry file
		}
	}

	return protocolIDMap[registry][namespacedID]
}

func NamespacedID(registry string, protocolID int32) string {
	if namespacedIDMap[registry] == nil {
		if err := readRegistry(registry); err != nil {
			panic(err) // failed to open registry file
		}
	}

	return namespacedIDMap[registry][protocolID]
}

func readRegistry(registryName string) error {
	registryJson := make(map[string]map[string]interface{})
	// Open our jsonFile
	jsonFile, err := os.Open("./data-generator/generated/reports/registries.json")
	if err != nil {
		return err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteValue, &registryJson)
	if err != nil {
		return err
	}

	entries := registryJson[registryName]["entries"].(map[string]interface{})
	if protocolIDMap[registryName] == nil {
		protocolIDMap[registryName] = make(map[string]int32, len(entries))
	}

	if namespacedIDMap[registryName] == nil {
		namespacedIDMap[registryName] = make(map[int32]string, len(entries))
	}

	for namespacedID, value := range entries {
		protocolID := int32(value.(map[string]interface{})["protocol_id"].(float64))
		protocolIDMap[registryName][namespacedID] = protocolID
		namespacedIDMap[registryName][protocolID] = namespacedID
	}

	return nil
}
