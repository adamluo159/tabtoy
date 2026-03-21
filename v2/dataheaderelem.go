package v2

import (
	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
	"strings"
)

type DataHeaderElement struct {
	FieldName string
	FieldType string
	FieldMeta string
	Comment   string
}

func checkElement(def *model.FieldDescriptor) int {
	return -1
}

func checkSameNameElement(exist, def *model.FieldDescriptor) int {
	// 多个同名字段只允许repeated方式的字段
	if !exist.IsRepeated {
		log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_DuplicateFieldName), def.Name)
		return DataSheetHeader_FieldName
	}

	// 多个repeated描述类型不一致
	if exist.Type != def.Type {

		log.Errorf("%s '%s' '%s' '%s'", i18n.String(i18n.DataHeader_RepeatedFieldTypeNotSameInMultiColumn),
			def.Name,
			model.FieldTypeToString(exist.Type),
			model.FieldTypeToString(def.Type))

		return DataSheetHeader_FieldType
	}

	// 多个repeated描述内建类型不一致
	if exist.Complex != def.Complex {

		log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_RepeatedFieldTypeNotSameInMultiColumn),
			def.Name)

		return DataSheetHeader_FieldType
	}

	// 多个repeated描述的meta不一致
	if exist.Meta.String() != def.Meta.String() {

		log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_RepeatedFieldMetaNotSameInMultiColumn),
			def.Name)

		return DataSheetHeader_FieldMeta
	}

	return -1
}

// parseStructPathField 解析 StructName.FieldName 格式的字段名
// 返回：结构体字段名、结构体内部字段名、是否成功
func parseStructPathField(fieldName string) (structPath, structFieldName string, ok bool) {
	dotIndex := strings.Index(fieldName, ".")
	if dotIndex == -1 {
		return "", "", false
	}
	structPath = fieldName[:dotIndex]
	structFieldName = fieldName[dotIndex+1:]
	if structPath == "" || structFieldName == "" {
		return "", "", false
	}
	return structPath, structFieldName, true
}

func (self *DataHeaderElement) Parse(def *model.FieldDescriptor, localFD *model.FileDescriptor, globalFD *model.FileDescriptor, headerByName map[string]*model.FieldDescriptor) int {

	// ====================解析字段名====================
	// 检查是否是 StructName.FieldName 格式
	if structPath, structFieldName, ok := parseStructPathField(self.FieldName); ok {
		def.StructPath = structPath
		def.StructFieldName = structFieldName
		def.Name = structPath // 主字段名使用结构体路径

		// 对于多列展开格式，检查是否已有主字段
		if existMain, ok := headerByName[structPath]; ok {
			// 已有主字段，从主字段获取类型信息
			def.Type = existMain.Type
			def.Complex = existMain.Complex
			def.IsRepeated = false // 子字段不是 repeated

			// 查找结构体字段类型
			for _, f := range def.Complex.Fields {
				if f.Name == structFieldName {
					// 设置子字段的类型
					def.Type = f.Type
					def.Complex = f.Complex
					break
				}
			}

			// 解析特性
			if err := def.Meta.Parse(self.FieldMeta); err != nil {
				log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_MetaParseFailed), err)
				return DataSheetHeader_FieldMeta
			}

			def.Comment = strings.Replace(self.Comment, "\n", " ", -1)
			return -1
		}

		// 没有主字段，需要解析类型
	} else {
		def.Name = self.FieldName
	}

	// ====================解析类型====================

	testFileD := localFD

	for {

		if def.ParseType(testFileD, self.FieldType) {
			break
		}

		if testFileD == localFD {
			testFileD = globalFD
			continue
		}

		break
	}

	// 依然找不到, 尝试延迟解析map类型
	if def.Type == model.FieldType_None && strings.HasPrefix(self.FieldType, "map<") {
		// 标记为map类型，稍后在延迟解析中处理
		def.Type = model.FieldType_Map
		def.RawFieldType = self.FieldType
	} else if def.Type == model.FieldType_None && def.StructPath != "" {
		// 多列展开格式，类型为空，稍后在 addStructExpandField 中处理
		// 这里暂时跳过
	} else if def.Type == model.FieldType_None {
		// 依然找不到, 报错
		log.Errorf("%s, '%s' (%s) raw: %s", i18n.String(i18n.DataHeader_TypeNotFound), def.Name, model.FieldTypeToString(def.Type), self.FieldType)
		return DataSheetHeader_FieldType
	}

	// ====================解析特性====================
	if err := def.Meta.Parse(self.FieldMeta); err != nil {
		log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_MetaParseFailed), err)
		return DataSheetHeader_FieldMeta
	}

	def.Comment = strings.Replace(self.Comment, "\n", " ", -1)

	// 多列展开的字段不进行同名字段检查
	if def.StructPath != "" {
		return -1
	}

	// 根据字段名查找, 处理repeated字段case
	exist, ok := headerByName[def.Name]

	if ok {

		// 多个同名字段只允许repeated方式的字段
		if !exist.IsRepeated {
			log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_DuplicateFieldName), def.Name)
			return DataSheetHeader_FieldName
		}

		// 多个repeated描述类型不一致
		if exist.Type != def.Type {

			log.Errorf("%s '%s' '%s' '%s'", i18n.String(i18n.DataHeader_RepeatedFieldTypeNotSameInMultiColumn),
				def.Name,
				model.FieldTypeToString(exist.Type),
				model.FieldTypeToString(def.Type))

			return DataSheetHeader_FieldType
		}

		// 多个repeated描述内建类型不一致
		if exist.Complex != def.Complex {

			log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_RepeatedFieldTypeNotSameInMultiColumn),
				def.Name)

			return DataSheetHeader_FieldType
		}

		// 多个repeated描述的meta不一致
		if exist.Meta.String() != def.Meta.String() {

			log.Errorf("%s '%s'", i18n.String(i18n.DataHeader_RepeatedFieldMetaNotSameInMultiColumn),
				def.Name)

			return DataSheetHeader_FieldMeta
		}

		def = exist
	}

	return -1
}
