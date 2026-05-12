package filter

import (
	"fmt"
	"strconv"

	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
)

// 从单元格原始数据到最终输出的数值, 检查并转换, 处理默认值及根据meta转换情况
func ConvertValue(fd *model.FieldDescriptor, value string, fileD *model.FileDescriptor, node *model.Node) (ret string, ok bool) {

	// 空格, 且有默认值时, 使用默认值
	if value == "" {
		value = fd.DefaultValue()
	}

	switch fd.Type {
	case model.FieldType_Int32:
		_, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			// 尝试在全局枚举中查找对应的别名
			if enumValue := findEnumValueByAlias(value, fileD); enumValue != "" {
				ret = enumValue
				node.AddValue(ret)
				break
			}
			log.Debugln(err)
			return "", false
		}

		ret = value
		node.AddValue(ret)
	case model.FieldType_Int64:
		_, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			// 尝试在全局枚举中查找对应的别名
			if enumValue := findEnumValueByAlias(value, fileD); enumValue != "" {
				ret = enumValue
				node.AddValue(ret)
				break
			}
			log.Debugln(err)
			return "", false
		}

		ret = value
		node.AddValue(ret)
	case model.FieldType_UInt32:
		_, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			// 尝试在全局枚举中查找对应的别名
			if enumValue := findEnumValueByAlias(value, fileD); enumValue != "" {
				ret = enumValue
				node.AddValue(ret)
				break
			}
			log.Debugln(err)
			return "", false
		}

		ret = value
		node.AddValue(ret)
	case model.FieldType_UInt64:
		_, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			// 尝试在全局枚举中查找对应的别名
			if enumValue := findEnumValueByAlias(value, fileD); enumValue != "" {
				ret = enumValue
				node.AddValue(ret)
				break
			}
			log.Debugln(err)
			return "", false
		}

		ret = value
		node.AddValue(ret)
	case model.FieldType_Float:
		_, err := strconv.ParseFloat(value, 32)
		if err != nil {
			log.Debugln(err)
			return "", false
		}

		ret = value
		node.AddValue(ret)
	case model.FieldType_Bool:

		for {
			if value == "是" {
				ret = "true"
				break
			} else if value == "否" {
				ret = "false"
				break
			}

			v, err := strconv.ParseBool(value)

			if err != nil {
				log.Debugln(err)
				return "", false
			}

			if v {
				ret = "true"
			} else {
				ret = "false"
			}

			break
		}

		node.AddValue(ret)

	case model.FieldType_String:
		ret = value
		node.AddValue(ret)
	case model.FieldType_Enum:
		if fd.Complex == nil {
			log.Errorf("%s, '%s'", i18n.String(i18n.ConvertValue_EnumTypeNil), fd.Name)
			return "", false
		}

		evd := fd.Complex.FieldByValueAndMeta(value)
		if evd == nil {
			log.Errorf("%s, '%s' '%s'", i18n.String(i18n.ConvertValue_EnumValueNotFound), value, fd.Complex.Name)
			return "", false
		}

		// 使用枚举的英文字段名输出
		ret = evd.Name
		node.AddValue(ret).EnumValue = evd.EnumValue

	case model.FieldType_Struct:

		if fd.Complex == nil {
			log.Errorf("%s, '%s'", i18n.String(i18n.ConvertValue_StructTypeNil), fd.Name)
			return "", false
		}

		if value == "" {

			if !fillStructDefaultValue(fd.Complex, fileD, node) {
				return "", false
			}

		} else {
			if !parseStruct(fd, value, fileD, node) {
				return "", false
			}
		}

	case model.FieldType_Map:

		if value == "" {
			return "", true
		}

		if !parseMapValue(fd, value, fileD, node) {
			return "", false
		}

	default:
		log.Errorf("%s, '%s' '%s'", i18n.String(i18n.ConvertValue_UnknownFieldType), fd.Name, fd.Name)
		return "", false
	}

	ok = true

	return
}

// 填充空结构体的默认值
func fillStructDefaultValue(structD *model.Descriptor, fileD *model.FileDescriptor, node *model.Node) bool {

	for _, fd := range structD.Fields {

		// 没默认值不输出, 建议忽略的字段除外, 先导出node, 再在printer中忽略
		if fd.Meta.GetString("Default") == "" {
			if node.SugguestIgnore {
				continue
			}
			// 没有默认值且不需要忽略，创建一个标记为忽略的节点
			fieldNode := node.AddKey(fd)
			fieldNode.SugguestIgnore = true
			continue
		}

		fieldNode := node.AddKey(fd)

		_, ok := ConvertValue(fd, "", fileD, fieldNode)
		if !ok {
			return false
		}
	}

	return true

}

func parseMapValue(fd *model.FieldDescriptor, value string, fileD *model.FileDescriptor, node *model.Node) bool {
	spliter := fd.MapSpliter()
	if spliter == "" {
		spliter = "|"
	}

	entries := splitMapEntries(value, spliter)

	keySet := make(map[string]bool)

	for _, entry := range entries {
		entry = trimString(entry)
		if entry == "" {
			continue
		}

		entryNode := node.AddKey(fd)

		if fd.MapValueType == model.FieldType_Struct {
			if entry == "" {
				if fd.MapValueComplex == nil {
					log.Errorf("map value struct type is nil for field: '%s'", fd.Name)
					return false
				}
				if !fillStructDefaultValue(fd.MapValueComplex, fileD, entryNode) {
					return false
				}
			} else {
				keyStr, err := parseMapStructValueAndGetKey(fd, entry, fileD, entryNode)
				if err != nil {
					log.Errorf("%s", err.Error())
					return false
				}

				if keySet[keyStr] {
					log.Errorf("duplicate map key: '%s' in field: '%s'", keyStr, fd.Name)
					return false
				}
				keySet[keyStr] = true
				entryNode.MapKey = keyStr
			}
		} else {
			colonIndex := findColonIndex(entry)
			if colonIndex == -1 {
				log.Errorf("map entry format error, expected 'key:value', got: '%s'", entry)
				return false
			}

			keyStr := trimString(entry[:colonIndex])
			valueStr := trimString(entry[colonIndex+1:])

			// Handle enum key type
			if fd.MapKeyComplex != nil {
				keyEvd := fd.MapKeyComplex.FieldByValueAndMeta(keyStr)
				if keyEvd == nil {
					log.Errorf("enum key value not found: '%s' in enum '%s'", keyStr, fd.MapKeyComplex.Name)
					return false
				}
				keyStr = keyEvd.Name
				entryNode.EnumKey = keyEvd.EnumValue
			}

			if keySet[keyStr] {
				log.Errorf("duplicate map key: '%s' in field: '%s'", keyStr, fd.Name)
				return false
			}
			keySet[keyStr] = true
			entryNode.MapKey = keyStr

			valueNode := entryNode.AddValue(valueStr)
			if fd.MapValueType == model.FieldType_Enum {
				if fd.MapValueComplex == nil {
					log.Errorf("map value enum type is nil for field: '%s'", fd.Name)
					return false
				}
				evd := fd.MapValueComplex.FieldByValueAndMeta(valueStr)
				if evd == nil {
					log.Errorf("enum value not found: '%s' in enum '%s'", valueStr, fd.MapValueComplex.Name)
					return false
				}
				valueNode.EnumValue = evd.EnumValue
			}
		}
	}

	return true
}

func parseMapStructValueAndGetKey(fd *model.FieldDescriptor, value string, fileD *model.FileDescriptor, node *model.Node) (string, error) {
	if fd.MapValueComplex == nil {
		return "", fmt.Errorf("map value struct type is nil for field: '%s'", fd.Name)
	}

	mapKeyField := fd.MapKeyField()
	if mapKeyField == "" {
		return "", fmt.Errorf("MapKeyField is required when map value is struct, field: '%s'", fd.Name)
	}

	keyFieldDef := fd.MapValueComplex.FieldByValueAndMeta(mapKeyField)
	if keyFieldDef == nil {
		return "", fmt.Errorf("MapKeyField '%s' not found in struct '%s'", mapKeyField, fd.MapValueComplex.Name)
	}

	p := newStructParser(value)

	sfList := newStructFieldList()

	// 创建一个临时的 FieldDescriptor 用于结构体解析
	tempFd := &model.FieldDescriptor{
		Complex: fd.MapValueComplex,
	}

	result := p.Run(tempFd, func(k, v string) bool {
		bnField := fd.MapValueComplex.FieldByValueAndMeta(k)
		if bnField == nil {
			log.Errorf("struct field not found: '%s'", k)
			return false
		}

		if sfList.Exists(bnField) {
			log.Errorf("duplicate field in cell: '%s'", k)
			return false
		}

		sfList.Add(bnField, v)
		return true
	})

	if !result {
		return "", fmt.Errorf("parse struct failed")
	}

	keyFieldValue := ""
	if sfList.Exists(keyFieldDef) {
		keyFieldValue = sfList.GetValue(keyFieldDef)
	} else {
		return "", fmt.Errorf("MapKeyField '%s' value not found in cell", mapKeyField)
	}

	// Handle enum key type
	if fd.MapKeyComplex != nil {
		keyEvd := fd.MapKeyComplex.FieldByValueAndMeta(keyFieldValue)
		if keyEvd == nil {
			return "", fmt.Errorf("enum key value not found: '%s' in enum '%s'", keyFieldValue, fd.MapKeyComplex.Name)
		}
		node.EnumKey = keyEvd.EnumValue
		keyFieldValue = keyEvd.Name
	}

	for _, structField := range fd.MapValueComplex.Fields {
		if sfList.Exists(structField) {
			continue
		}

		if structField.Meta.GetString("Default") != "" {
			sfList.Add(structField, structField.Meta.GetString("Default"))
		}
	}

	sfList.Sort()

	for i := 0; i < sfList.Len(); i++ {
		v := sfList.Get(i)
		fieldNode := node.AddKey(v.key)
		_, ok := ConvertValue(v.key, v.value, fileD, fieldNode)
		if !ok {
			return "", fmt.Errorf("convert value failed for field '%s'", v.key.Name)
		}
	}

	return keyFieldValue, nil
}

func splitMapEntries(value, spliter string) []string {
	var entries []string
	var current string
	inString := false

	for _, c := range value {
		if c == '"' {
			inString = !inString
			current += string(c)
		} else if c == rune(spliter[0]) && !inString && len(spliter) == 1 {
			entries = append(entries, current)
			current = ""
		} else {
			current += string(c)
		}
	}

	if current != "" {
		entries = append(entries, current)
	}

	return entries
}

func findColonIndex(s string) int {
	inString := false
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			inString = !inString
		} else if s[i] == ':' && !inString {
			return i
		}
	}
	return -1
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// findEnumValueByAlias 在全局枚举中查找对应的别名，返回枚举值
// 当 int32/int64/uint32/uint64 字段的值无法解析为数字时调用
// 例如：value="金币" 可能对应 ConstItemID 枚举中的值 "10000"
func findEnumValueByAlias(value string, fileD *model.FileDescriptor) string {
	if fileD == nil {
		return ""
	}

	// 遍历所有描述符，查找枚举类型
	for _, desc := range fileD.Descriptors {
		if desc.Kind != model.DescriptorKind_Enum {
			continue
		}

		// 在枚举中查找匹配的字段（按名称或别名）
		field := desc.FieldByValueAndMeta(value)
		if field != nil {
			// 返回枚举值（数字字符串）
			return fmt.Sprintf("%d", field.EnumValue)
		}
	}

	return ""
}
