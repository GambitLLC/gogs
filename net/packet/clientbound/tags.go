package clientbound

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gogs/data"
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
	"io/ioutil"
	"log"
	"os"
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
		BlockTags:  getTagsArray("blocks", "minecraft:block"),
		ItemTags:   getTagsArray("items", "minecraft:item"),
		FluidTags:  getTagsArray("fluids", "minecraft:fluid"),
		EntityTags: getTagsArray("entity_types", "minecraft:entity_type"),
	}

	return vanillaTags
}

func getTagsArray(tag string, registry string) TagsArray {
	arr := make(TagsArray, 0, 64)

	files, err := ioutil.ReadDir(fmt.Sprintf("./data-generator/generated/data/minecraft/tags/%s", tag))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		filename := f.Name()
		filename = strings.TrimSuffix(filename, filepath.Ext(filename))

		arr = append(arr, Tag{
			TagName: pk.Identifier(filename),
			Entries: getEntries(tag, registry, filename),
		})
	}

	return arr
}

func getEntries(rootTag string, registry string, tag string) []pk.VarInt {
	file, _ := os.Open(fmt.Sprintf("./data-generator/generated/data/minecraft/tags/%s/%s.json", rootTag, tag))
	byteValue, _ := ioutil.ReadAll(file)
	valueMap := make(map[string][]string)
	_ = json.Unmarshal(byteValue, &valueMap)

	entries := make([]pk.VarInt, 0, len(valueMap["values"]))
	for _, v := range valueMap["values"] {
		if v[0] == '#' {
			entries = append(entries, getEntries(rootTag, registry, v[11:])...) // trim "#minecraft:"
		} else {
			entries = append(entries, pk.VarInt(data.ProtocolID(registry, v)))
		}
	}
	return entries
}
