package v2

import (
	"strings"

	"github.com/adamluo159/tabtoy/v2/filter"
	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
)

func coloumnProcessor(file model.GlobalChecker, record *model.Record, fd *model.FieldDescriptor, raw string, sugguestIgnore bool) bool {

	spliter := fd.ListSpliter()

	if fd.IsRepeated && spliter != "" {

		valueList := strings.Split(raw, spliter)

		var node *model.Node

		if fd.Type != model.FieldType_Struct {
			node = record.NewNodeByDefine(fd)
		}

		for _, v := range valueList {

			rawSingle := strings.TrimSpace(v)

			// 结构体要多添加一个节点, 处理repeated 结构体情况
			if fd.Type == model.FieldType_Struct {
				node = record.NewNodeByDefine(fd)
				node.StructRoot = true
				node = node.AddKey(fd)
			}

			if raw != "" {
				if !dataProcessor(file, fd, rawSingle, node) {
					return false
				}
			}

		}

	} else { // 普通数据/repeated单元格分多个列

		node := record.NewNodeByDefine(fd)

		node.SugguestIgnore = sugguestIgnore

		// 结构体要多添加一个节点, 处理repeated 结构体情况
		if fd.Type == model.FieldType_Struct {

			node.StructRoot = true
			node = node.AddKey(fd)
		}

		node.SugguestIgnore = sugguestIgnore

		if !dataProcessor(file, fd, raw, node) {
			return false
		}
	}

	return true
}

// structExpandProcessor 处理多列展开的结构体字段
func structExpandProcessor(file model.GlobalChecker, record *model.Record, fv *model.FieldValue) bool {
	info, ok := fv.StructExpandInfo.(*StructExpandInfo)
	if !ok {
		log.Errorf("invalid struct expand info")
		return false
	}

	// 获取或创建主字段节点
	mainNode := record.NewNodeByDefine(fv.FieldDef)
	mainNode.IsRepeated = true

	// 计算需要创建多少个结构体实例
	maxInstanceIndex := 0
	for _, col := range info.Columns {
		if col.StructFieldIndex > maxInstanceIndex {
			maxInstanceIndex = col.StructFieldIndex
		}
	}

	// 确保有足够多的结构体实例节点
	for len(mainNode.Child) <= fv.StructInstanceIndex {
		structNode := &model.Node{
			FieldDescriptor: fv.FieldDef,
			StructRoot:      true,
		}
		mainNode.Child = append(mainNode.Child, structNode)
	}

	// 获取当前结构体实例节点
	structNode := mainNode.Child[fv.StructInstanceIndex]

	// 添加字段节点
	fieldNode := structNode.AddKey(fv.StructFieldDef)

	// 转换值
	if !dataProcessor(file, fv.StructFieldDef, fv.RawValue, fieldNode) {
		return false
	}

	return true
}

func dataProcessor(gc model.GlobalChecker, fd *model.FieldDescriptor, raw string, node *model.Node) bool {

	// 单值
	if cv, ok := filter.ConvertValue(fd, raw, gc.GlobalFileDesc(), node); !ok {
		goto ConvertError

	} else {

		// 值重复检查
		if fd.Meta.GetBool("RepeatCheck") && !gc.CheckValueRepeat(fd, cv) {
			log.Errorf("%s, %s raw: '%s'", i18n.String(i18n.DataSheet_ValueRepeated), fd.String(), cv)
			return false
		}
	}

	return true

ConvertError:

	log.Errorf("%s, %s raw: '%s'", i18n.String(i18n.DataSheet_ValueConvertError), fd.String(), raw)

	return false
}
