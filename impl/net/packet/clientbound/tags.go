package clientbound

import (
	"bytes"
	"fmt"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type Tags struct {
	BlockTags  TagsArray
	ItemTags   TagsArray
	FluidTags  TagsArray
	EntityTags TagsArray
}

func (s Tags) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.Tags, s.BlockTags, s.ItemTags, s.FluidTags, s.EntityTags)
}

type TagsArray []Tag

func (a TagsArray) Encode() []byte {
	buf := bytes.Buffer{}

	buf.Write(pk.VarInt(len(a)).Encode())
	for _, v := range a {
		buf.Write(v.Encode())
	}

	return buf.Bytes()
}

type Tag struct {
	TagName pk.Identifier
	Entries []pk.VarInt
}

func (s Tag) Encode() []byte {
	buf := bytes.Buffer{}

	buf.Write(s.TagName.Encode())
	buf.Write(pk.VarInt(len(s.Entries)).Encode())
	for _, v := range s.Entries {
		buf.Write(v.Encode())
	}

	return buf.Bytes()
}

var vanillaTags Tags

func VanillaTags() Tags {
	if vanillaTags.BlockTags != nil {
		return vanillaTags
	}

	vanillaTags = Tags{
		//BlockTags:  getTags("blocks", "minecraft:block"),
		//ItemTags:   getTags("items", "minecraft:item"),
		//FluidTags:  getTags("fluids", "minecraft:fluid"),
		//EntityTags: getTags("entity_types", "minecraft:entity_type"),
		BlockTags:  getTags("blocks"),
		ItemTags:   getTags("items"),
		FluidTags:  getTags("fluids"),
		EntityTags: getTags("entity_types"),
	}

	return vanillaTags
}

// func getTags(tag string, registry string) TagsArray {
func getTags(tag string) TagsArray {
	arr := make(TagsArray, 0, 64)

	dir := fmt.Sprintf("./data-generator/generated/data/minecraft/tags/%s", tag)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		filename := f.Name()
		filename = strings.TrimSuffix(filename, filepath.Ext(filename))

		/*
			// TODO: Populate entries using the registry (some tags reference other tags, needs recursion
			file, _ := os.Open(fmt.Sprintf("%s/%s", dir, f.Name()))
			byteValue, _ := ioutil.ReadAll(file)
			valueMap := make(map[string][]string)
			_ = json.Unmarshal(byteValue, &valueMap)

			entries := make([]pk.VarInt, len(valueMap["values"]))
			for i, v := range valueMap["values"] {
				entries[i] = pk.VarInt(data.RegistryID(registry, v))
			}
		*/

		if filename == "lava" {
			arr = append(arr, Tag{
				TagName: pk.Identifier(filename),
				Entries: []pk.VarInt{3, 4},
			})
		} else {
			arr = append(arr, Tag{
				TagName: pk.Identifier(filename),
				Entries: nil,
			})
		}

	}

	return arr
}
