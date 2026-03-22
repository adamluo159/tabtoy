package printer

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/adamluo159/tabtoy/util"
	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
)

func valueWrapperJson(t model.FieldType, node *model.Node) string {

	switch t {
	case model.FieldType_String:
		return util.StringWrap(util.StringEscape(node.Value))
	case model.FieldType_Enum:
		return strconv.Itoa(int(node.EnumValue))
	}

	return node.Value
}

func mapKeyWrapperJson(parent *model.Node, entryNode *model.Node) string {
	switch parent.MapKeyType {
	case model.FieldType_String:
		return util.StringWrap(util.StringEscape(entryNode.MapKey))
	case model.FieldType_Enum:
		return util.StringWrap(strconv.Itoa(int(entryNode.EnumKey)))
	default:
		return util.StringWrap(entryNode.MapKey)
	}
}

func printMapJson(bf *Stream, node *model.Node) {
	bf.Printf("{ ")
	for entryIndex, entryNode := range node.Child {
		if entryIndex > 0 {
			bf.Printf(", ")
		}

		bf.Printf("%s: ", mapKeyWrapperJson(node, entryNode))

		if node.MapValueType == model.FieldType_Struct {
			bf.Printf("{ ")
			var hasWriteField bool
			for fieldIndex, fieldNode := range entryNode.Child {
				if fieldNode.SugguestIgnore {
					continue
				}
				if hasWriteField && fieldIndex > 0 {
					bf.Printf(", ")
					hasWriteField = false
				}
				if len(fieldNode.Child) == 0 {
					log.Errorf("json printer: field node has no child, field: %s", fieldNode.Name)
					continue
				}
				valueNode := fieldNode.Child[0]
				bf.Printf("\"%s\": %s", fieldNode.Name, valueWrapperJson(fieldNode.Type, valueNode))
				hasWriteField = true
			}
			bf.Printf(" }")
		} else {
			valueNode := entryNode.Child[0]
			bf.Printf("%s", valueWrapperJson(node.MapValueType, valueNode))
		}
	}
	bf.Printf(" }")
}

type jsonPrinter struct {
}

func (self *jsonPrinter) Run(g *Globals) *Stream {
	if g.JsonDir != "" {
		if err := os.MkdirAll(g.JsonDir, 0755); err != nil {
			log.Errorf("create json dir failed: %v", err)
			return nil
		}

		for _, tab := range g.Tables {
			if !tab.LocalFD.MatchTag(".json") {
				log.Infof("%s: %s", i18n.String(i18n.Printer_IgnoredByOutputTag), tab.Name())
				continue
			}
			if tab.LocalFD.Name == "Map" || tab.LocalFD.Name == "Scene" {
				log.Infof("%s:不转数据，只用表结构", tab.Name())
				continue
			}

			bf := NewStream()
			bf.Printf("[")

			for rIndex, r := range tab.Recs {
				if rIndex > 0 {
					bf.Printf(",")
				}
				bf.Printf("\n\t{")

				var hasWriteColumn bool

				for rootFieldIndex, node := range r.Nodes {
					if node.SugguestIgnore {
						continue
					}

					if hasWriteColumn && rootFieldIndex > 0 {
						bf.Printf(", ")
						hasWriteColumn = false
					}

					if node.IsRepeated {
						bf.Printf("\"%s\":[ ", node.Name)
					} else if node.Type == model.FieldType_Map {
						bf.Printf("\"%s\": ", node.Name)
					} else {
						bf.Printf("\"%s\": ", node.Name)
					}

					if node.Type == model.FieldType_Map {
						printMapJson(bf, node)
					} else if node.Type != model.FieldType_Struct {
						if node.IsRepeated {
							for arrIndex, valueNode := range node.Child {
								bf.Printf("%s", valueWrapperJson(node.Type, valueNode))
								if arrIndex < len(node.Child)-1 {
									bf.Printf(", ")
								}
							}
						} else {
							valueNode := node.Child[0]
							bf.Printf("%s", valueWrapperJson(node.Type, valueNode))
						}
					} else {
						for structIndex, structNode := range node.Child {
							bf.Printf("{")

							var hasWriteField bool

							for structFieldIndex, fieldNode := range structNode.Child {
								if fieldNode.SugguestIgnore {
									continue
								}

								if hasWriteField && structFieldIndex > 0 {
									bf.Printf(", ")
									hasWriteField = false
								}

								if len(fieldNode.Child) == 0 {
									log.Errorf("json printer: field node has no child, field: %s", fieldNode.Name)
									continue
								}

								valueNode := fieldNode.Child[0]
								bf.Printf("\"%s\": %s", fieldNode.Name, valueWrapperJson(fieldNode.Type, valueNode))
								hasWriteField = true
							}

							bf.Printf(" }")

							if structIndex < len(node.Child)-1 {
								bf.Printf(", ")
							}
						}
					}

					if node.IsRepeated {
						bf.Printf(" ]")
					}

					hasWriteColumn = true
				}

				bf.Printf(" }")
			}

			bf.Printf("\n]")

			outputPath := filepath.Join(g.JsonDir, tab.LocalFD.Name+".json")
			if err := bf.WriteFile(outputPath); err != nil {
				log.Errorf("write json file failed: %v", err)
				return nil
			}

			log.Infof("[json] %s", outputPath)
		}

		return NewStream()
	}

	bf := NewStream()
	bf.Printf("{\n")

	for tabIndex, tab := range g.Tables {
		if !tab.LocalFD.MatchTag(".json") {
			log.Infof("%s: %s", i18n.String(i18n.Printer_IgnoredByOutputTag), tab.Name())
			continue
		}
		if tab.LocalFD.Name == "Map" || tab.LocalFD.Name == "Scene" {
			log.Infof("%s:不转数据，只用表结构", tab.Name())
			continue
		}

		if tabIndex > 0 {
			bf.Printf(", \n")
		}

		if !printTableJson(bf, tab) {
			return nil
		}
	}
	if g.TiledFileDir != "" {
		WriteTiledData(g, bf, g.TiledFileDir)
	}

	bf.Printf("}")

	return bf
}

func printTableJson(bf *Stream, tab *model.Table) bool {

	bf.Printf("	\"%s\":[\n", tab.LocalFD.Name)

	for rIndex, r := range tab.Recs {

		bf.Printf("		{ ")

		var hasWriteColumn bool

		for rootFieldIndex, node := range r.Nodes {

			if node.SugguestIgnore {
				continue
			}

			if hasWriteColumn && rootFieldIndex > 0 {
				bf.Printf(", ")
				hasWriteColumn = false
			}

			if node.IsRepeated {
				bf.Printf("\"%s\":[ ", node.Name)
			} else if node.Type == model.FieldType_Map {
				bf.Printf("\"%s\": ", node.Name)
			} else {
				bf.Printf("\"%s\": ", node.Name)
			}

			if node.Type == model.FieldType_Map {
				printMapJson(bf, node)
			} else if node.Type != model.FieldType_Struct {

				if node.IsRepeated {

					for arrIndex, valueNode := range node.Child {

						bf.Printf("%s", valueWrapperJson(node.Type, valueNode))

						if arrIndex < len(node.Child)-1 {
							bf.Printf(", ")
						}

					}
				} else {
					valueNode := node.Child[0]

					bf.Printf("%s", valueWrapperJson(node.Type, valueNode))

				}

			} else {

				for structIndex, structNode := range node.Child {

					bf.Printf("{ ")

					var hasWriteField bool

					for structFieldIndex, fieldNode := range structNode.Child {

						if fieldNode.SugguestIgnore {
							continue
						}

						if hasWriteField && structFieldIndex > 0 {
							bf.Printf(", ")
							hasWriteField = false
						}

						if len(fieldNode.Child) == 0 {
							log.Errorf("json printer: field node has no child, field: %s", fieldNode.Name)
							continue
						}

						valueNode := fieldNode.Child[0]

						bf.Printf("\"%s\": %s", fieldNode.Name, valueWrapperJson(fieldNode.Type, valueNode))

						hasWriteField = true
					}

					bf.Printf(" }")

					if structIndex < len(node.Child)-1 {
						bf.Printf(", ")
					}

				}

			}

			if node.IsRepeated {
				bf.Printf(" ]")
			}

			hasWriteColumn = true

		}

		bf.Printf(" }")

		if rIndex < len(tab.Recs)-1 {
			bf.Printf(",")
		}

		bf.Printf("\n")

	}

	bf.Printf("	]")

	return true

}

func init() {

	RegisterPrinter("json", &jsonPrinter{})

}
