package printer

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

type TiledTmx struct {
	XMLName        xml.Name `xml:"map"`
	Text           string   `xml:",chardata"`
	Version        string   `xml:"version,attr"`
	Tiledversion   string   `xml:"tiledversion,attr"`
	Orientation    string   `xml:"orientation,attr"`
	Renderorder    string   `xml:"renderorder,attr"`
	Width          string   `xml:"width,attr"`
	Height         string   `xml:"height,attr"`
	Tilewidth      string   `xml:"tilewidth,attr"`
	Tileheight     string   `xml:"tileheight,attr"`
	Infinite       string   `xml:"infinite,attr"`
	Nextlayerid    string   `xml:"nextlayerid,attr"`
	Nextobjectid   string   `xml:"nextobjectid,attr"`
	Editorsettings struct {
		Text   string `xml:",chardata"`
		Export struct {
			Text   string `xml:",chardata"`
			Target string `xml:"target,attr"`
			Format string `xml:"format,attr"`
		} `xml:"export"`
	} `xml:"editorsettings"`
	Tileset struct {
		Text     string `xml:",chardata"`
		Firstgid string `xml:"firstgid,attr"`
		Source   string `xml:"source,attr"`
	} `xml:"tileset"`
	Imagelayer struct {
		Text  string `xml:",chardata"`
		ID    string `xml:"id,attr"`
		Name  string `xml:"name,attr"`
		Image struct {
			Text   string `xml:",chardata"`
			Source string `xml:"source,attr"`
			Width  string `xml:"width,attr"`
			Height string `xml:"height,attr"`
		} `xml:"image"`
	} `xml:"imagelayer"`
	Layer struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Name   string `xml:"name,attr"`
		Width  string `xml:"width,attr"`
		Height string `xml:"height,attr"`
		Data   struct {
			Text     string `xml:",chardata"`
			Encoding string `xml:"encoding,attr"`
		} `xml:"data"`
	} `xml:"layer"`
}

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
	ID         int32
	Width      int32
	Height     int32
	TileWidth  int32
	TileHeight int32
	XCount     int32
	YCount     int32
	Tiles      []int32
}

func (m *GameMap) decodeCSV(raw string, tileTypeMap map[int32]string) (err error) {
	cleaner := func(r rune) rune {
		if (r >= '0' && r <= '9') || r == ',' {
			return r
		}
		return -1
	}
	rawDataClean := strings.Map(cleaner, raw)
	str := strings.Split(string(rawDataClean), ",")
	for _, s := range str {
		tile := cast.ToInt32(s) - 1
		if tileTypeMap[tile] == "" {
			log.Errorf("readTiledFile tileTypeMap[%d] is nil", tile)
			return fmt.Errorf("readTiledFile %+v tileTypeMap[%d] is nil", tileTypeMap, tile)
		}
		m.Tiles = append(m.Tiles, tile)
	}
	return err
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
	terrainMap := make(map[int32]string)
	des := g.DescriptorByName["TerrainType"]
	if des == nil {
		log.Errorln("TerrainType Descriptor nil")
		return
	}
	for _, v := range des.Fields {
		terrainMap[v.EnumValue] = v.Meta.KVPair.GetString("Alias")
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

func writeTileMap(path string, terrainMap map[int32]string) string {
	fbytes, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("readTiledFile path:%s err:%v", path, err)
		return ""
	}
	tmx := &TiledTmx{}
	err = xml.Unmarshal(fbytes, tmx)
	if err != nil {
		log.Errorf("readTiledFile umarshal data:%s err:%v", string(fbytes), err)
		return ""
	}

	fileName := filepath.Base(tmx.Imagelayer.Image.Source)
	gmap := &GameMap{
		ID:         cast.ToInt32(strings.Split(fileName, ".")[0]),
		Width:      cast.ToInt32(tmx.Width) * cast.ToInt32(tmx.Tilewidth),
		Height:     cast.ToInt32(tmx.Height) * cast.ToInt32(tmx.Tileheight),
		TileWidth:  cast.ToInt32(tmx.Tilewidth),
		TileHeight: cast.ToInt32(tmx.Tileheight),
		XCount:     cast.ToInt32(tmx.Width),
		YCount:     cast.ToInt32(tmx.Height),
	}
	err = gmap.decodeCSV(tmx.Layer.Data.Text, terrainMap)
	if err != nil {
		log.Errorf("readTiledFile decodeCSV path:%s error:%v", path, err)
		return ""
	}
	bys, err := json.Marshal(gmap)
	if err != nil {
		log.Errorf("readTiledFile json marshal path:%s error:%v", path, err)
		return ""
	}
	return string(bys)
}
