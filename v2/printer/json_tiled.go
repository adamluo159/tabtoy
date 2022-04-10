package printer

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type EditorMap struct {
	XMLName      xml.Name `xml:"tileset"`
	Text         string   `xml:",chardata"`
	Version      string   `xml:"version,attr"`
	Tiledversion string   `xml:"tiledversion,attr"`
	Name         string   `xml:"name,attr"`
	Tilewidth    string   `xml:"tilewidth,attr"`
	Tileheight   string   `xml:"tileheight,attr"`
	Tilecount    string   `xml:"tilecount,attr"`
	Columns      string   `xml:"columns,attr"`
	Image        struct {
		Source string `xml:"source,attr"`
	} `xml:"image"`

	Tile []struct {
		Text string `xml:",chardata"`
		ID   string `xml:"id,attr"`
		Type string `xml:"type,attr"`
	} `xml:"tile"`
}

type GameMap struct {
	ID        int32
	Name      string
	Width     int32
	Height    int32
	Xsize     int32
	HalfXsize int32
	Xcount    int32
	Ysize     int32
	HalfYsize int32
	Ycount    int32
	Tiles     []int32
}

func WriteTiledData(g *Globals, bf *Stream, patterns ...string) {
	matches := []string{}
	for _, pattern := range patterns {
		m, err := filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}
		matches = append(matches, m...)
	}
	if len(matches) == 0 {
		return
	}
	terrainMap := make(map[string]int32)
	des := g.DescriptorByName["TerrainType"]
	if des == nil {
		log.Errorln("TerrainType Descriptor nil")
		return
	}
	for _, v := range des.Fields {
		terrainMap[v.Meta.KVPair.GetString("Alias")] = v.EnumValue
	}
	bf.Printf(",\n")
	bf.Printf("	\"Map\":[\n")
	for i := 0; i < len(matches); i++ {
		match := matches[i]
		if s, e := os.Stat(match); e != nil || s.IsDir() {
			continue
		}
		data := writeTileMap(match, terrainMap)
		if data != "" {
			if len(matches)-1 != i {
				bf.Printf("		%s,\n", data)
			} else {
				bf.Printf("		%s\n", data)
			}
		}
	}
	bf.Printf("\t]")
}

func writeTileMap(path string, terrainMap map[string]int32) string {
	fbytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("readTiledFile path:%s err:%v", path, err)
		return ""
	}
	emap := &EditorMap{}
	err = xml.Unmarshal(fbytes, emap)
	if err != nil {
		err = json.Unmarshal(fbytes, emap)
	}
	if err != nil {
		log.Errorf("readTiledFile umarshal data:%s err:%v", string(fbytes), err)
		return ""
	}
	name := filepath.Base(emap.Image.Source)
	id, err := strconv.Atoi(strings.TrimSuffix(name, filepath.Ext(name)))
	if err != nil {
		log.Errorln("image  strconv.Atoi err:%v name:%s", err, name)
		return ""
	}
	gmap := &GameMap{
		ID:    int32(id),
		Name:  name,
		Xsize: 96,
		Ysize: 96,
	}
	tileCount, _ := strconv.Atoi(emap.Tilecount)
	columns, _ := strconv.Atoi(emap.Columns)
	for i := 0; i < len(emap.Tile); i++ {
		tile := emap.Tile[i]
		if tile.ID != fmt.Sprintf("%d", i) {
			log.Errorf("tile.ID:%s != i:%d please check terrain type not empty !", tile.ID, i)
			return ""
		}
		t, ok := terrainMap[tile.Type]
		if !ok {
			log.Errorf("grid:%d type:%s is not in %+v", i, tile.Type, terrainMap)
			return ""
		}
		gmap.Tiles = append(gmap.Tiles, t)
	}
	gmap.Xcount = int32(columns)
	gmap.Ycount = int32(tileCount / columns)
	gmap.HalfXsize = gmap.Xsize / 2
	gmap.HalfYsize = gmap.Ysize / 2
	gmap.Width = gmap.Xcount * gmap.Xsize
	gmap.Height = gmap.Ycount * gmap.Ysize
	bys, err := json.Marshal(gmap)
	if err != nil {
		log.Errorf("readTiledFile json marshal path:%s error:%v", path, err)
		return ""
	}
	return string(bys)
}
