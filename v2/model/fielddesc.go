package model

import (
	"fmt"
	"strings"
)

type FieldType int

const (
	FieldType_None   FieldType = 0
	FieldType_Int32  FieldType = 1
	FieldType_Int64  FieldType = 2
	FieldType_UInt32 FieldType = 3
	FieldType_UInt64 FieldType = 4
	FieldType_Float  FieldType = 5
	FieldType_String FieldType = 6
	FieldType_Bool   FieldType = 7
	FieldType_Enum   FieldType = 8
	FieldType_Struct FieldType = 9
	FieldType_Table  FieldType = 10 // 表格, 仅限二进制使用
	FieldType_Map    FieldType = 11 // map类型
)

// 一列的描述
type FieldDescriptor struct {
	Name string

	Type FieldType

	Complex *Descriptor // 复杂类型: 枚举或者结构体

	Order int32 // 在Descriptor中的顺序

	Meta *MetaInfo // 扩展字段

	IsRepeated bool

	EnumValue int32 // 枚举值

	Comment string // 注释

	Parent *Descriptor

	MapKeyType   FieldType   // map的key类型
	MapKeyComplex *Descriptor // map的key复杂类型(枚举)
	MapValueType FieldType   // map的value类型
	MapValueComplex *Descriptor // map的value复杂类型(枚举或结构体)

	RawFieldType string // 原始类型字符串，用于延迟解析

	// 多列展开支持 (StructName.FieldName 格式)
	StructPath      string // 结构体路径，如 "Costs"
	StructFieldName string // 结构体字段名，如 "ID"
	StructColIndex  int    // 在 repeated 结构体中的列索引（第几个实例）
}

func NewFieldDescriptor() *FieldDescriptor {
	return &FieldDescriptor{
		Meta: NewMetaInfo(),
	}
}

func (self *FieldDescriptor) Tag() int32 {

	return MakeTag(int32(self.Type), self.Order)
}

func MakeTag(t int32, order int32) int32 {
	return t<<16 | order
}

func (self *FieldDescriptor) Equal(fd *FieldDescriptor) bool {

	if self.Name != fd.Name {
		return false
	}

	if self.Type != fd.Type {
		return false
	}

	if self.Meta.String() != fd.Meta.String() {
		return false
	}

	if self.IsRepeated != fd.IsRepeated {
		return false
	}

	if self.EnumValue != fd.EnumValue {
		return false
	}

	if self.complexName() != fd.complexName() {
		return false
	}

	return true
}

func (self *FieldDescriptor) complexName() string {
	if self.Complex != nil {
		return self.Complex.Name
	}

	return ""
}

// 自动适配结构体和枚举输出合适的类型, 类型名为go only
func (self *FieldDescriptor) TsTypeString() string {
	if self.Complex != nil {
		return self.Complex.Name
	} else {
		return FieldTypeToTsString(self.Type)
	}
}

// 自动适配结构体和枚举输出合适的类型, 类型名为go only
func (self *FieldDescriptor) TypeString() string {
	if self.Complex != nil {
		return self.Complex.Name
	} else {
		return FieldTypeToString(self.Type)
	}
}

func (self *FieldDescriptor) KindString() string {
	return FieldTypeToString(self.Type)
}

func (self *FieldDescriptor) String() string {

	var repString string
	if self.IsRepeated {
		repString = "repeated "
	}

	return fmt.Sprintf("name: '%s' %stype: '%s'", self.Name, repString, self.TypeString())
}

// IsOneToManyIndex 判断是否是一对多索引（只有MakeIndex没有RepeatCheck）
func (self *FieldDescriptor) IsOneToManyIndex() bool {
	return self.Meta.GetBool("MakeIndex") && !self.Meta.GetBool("RepeatCheck")
}

func (self *FieldDescriptor) DefaultValue() string {

	if v := self.Meta.GetString("Default"); v != "" {
		return v
	}

	switch self.Type {
	case FieldType_Int32,
		FieldType_UInt32,
		FieldType_Int64,
		FieldType_UInt64,
		FieldType_Float:
		return "0"
	case FieldType_Bool:
		return "false"
	case FieldType_Enum:

		if self.Complex == nil {
			log.Debugln("build type null while get default value", self.Name)
			return ""
		}

		if len(self.Complex.Fields) == 0 {
			return ""
		}

		return self.Complex.Fields[0].Name

	}

	return ""
}

func (self *FieldDescriptor) ListSpliter() string {

	return self.Meta.GetString("ListSpliter")
}

func (self *FieldDescriptor) RepeatCheck() bool {

	return self.Meta.GetBool("RepeatCheck")
}

func (self *FieldDescriptor) MapSpliter() string {

	return self.Meta.GetString("MapSpliter")
}

func (self *FieldDescriptor) MapKeyField() string {

	return self.Meta.GetString("MapKeyField")
}

func (self *FieldDescriptor) IsMap() bool {

	return self.Type == FieldType_Map
}

var tsStrByFieldDescriptor = map[FieldType]string{
	FieldType_None:   "none",
	FieldType_Int32:  "number",
	FieldType_Int64:  "number",
	FieldType_UInt32: "number",
	FieldType_UInt64: "number",

	FieldType_Float:  "number",
	FieldType_String: "string",
	FieldType_Bool:   "bool",
	FieldType_Enum:   "enum",
	FieldType_Struct: "struct",
	FieldType_Map:    "map",
}

func FieldTypeToTsString(t FieldType) string {
	if v, ok := tsStrByFieldDescriptor[t]; ok {
		return v
	}

	return "unknown"
}

var strByFieldDescriptor = map[FieldType]string{
	FieldType_None:   "none",
	FieldType_Int32:  "int32",
	FieldType_Int64:  "int64",
	FieldType_UInt32: "uint32",
	FieldType_UInt64: "uint64",

	FieldType_Float:  "float",
	FieldType_String: "string",
	FieldType_Bool:   "bool",
	FieldType_Enum:   "enum",
	FieldType_Struct: "struct",
	FieldType_Map:    "map",
}

var fieldTypeByString = make(map[string]FieldType)

func FieldTypeToString(t FieldType) string {
	if v, ok := strByFieldDescriptor[t]; ok {
		return v
	}

	return "unknown"
}

func ParseFieldType(str string) (t FieldType, ok bool) {
	v, ok := fieldTypeByString[str]
	return v, ok
}

const RepeatedKeyword = "repeated"
const RepeatedKeywordLen = len(RepeatedKeyword)

const SliceKeyword = "[]"
const SliceKeywordLen = len(SliceKeyword)

const MapKeyword = "map<"

func (self *FieldDescriptor) ParseType(fileD *FileDescriptor, rawstr string) bool {

	var puretype string

	if strings.HasPrefix(rawstr, RepeatedKeyword) {

		puretype = rawstr[RepeatedKeywordLen+1:]

		self.IsRepeated = true
	} else if strings.HasPrefix(rawstr, SliceKeyword) {
		puretype = rawstr[SliceKeywordLen:]

		self.IsRepeated = true
	} else if strings.HasPrefix(rawstr, MapKeyword) {
		return self.parseMapType(fileD, rawstr)
	} else {
		puretype = rawstr
	}

	if ft, ok := ParseFieldType(puretype); ok {
		self.Type = ft
		return true
	}

	if desc, ok := fileD.DescriptorByName[puretype]; ok {
		self.Complex = desc

		switch desc.Kind {
		case DescriptorKind_Struct:
			self.Type = FieldType_Struct
		case DescriptorKind_Enum:
			self.Type = FieldType_Enum
		}

	} else {
		return false
	}

	return true
}

func (self *FieldDescriptor) parseMapType(fileD *FileDescriptor, rawstr string) bool {
	if !strings.HasSuffix(rawstr, ">") {
		return false
	}

	inner := rawstr[4 : len(rawstr)-1]
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		return false
	}

	keyType := strings.TrimSpace(parts[0])
	valueType := strings.TrimSpace(parts[1])

	if ft, ok := ParseFieldType(keyType); ok {
		self.MapKeyType = ft
	} else if desc, ok := fileD.DescriptorByName[keyType]; ok {
		self.MapKeyComplex = desc
		if desc.Kind == DescriptorKind_Enum {
			self.MapKeyType = FieldType_Enum
		} else {
			return false
		}
	} else {
		// key类型不存在，返回false，稍后延迟解析
		return false
	}

	if ft, ok := ParseFieldType(valueType); ok {
		self.MapValueType = ft
	} else if desc, ok := fileD.DescriptorByName[valueType]; ok {
		self.MapValueComplex = desc
		switch desc.Kind {
		case DescriptorKind_Struct:
			self.MapValueType = FieldType_Struct
		case DescriptorKind_Enum:
			self.MapValueType = FieldType_Enum
		}
	} else {
		// value类型不存在，返回false，稍后延迟解析
		return false
	}

	self.Type = FieldType_Map
	return true
}

func init() {

	for k, v := range strByFieldDescriptor {
		fieldTypeByString[v] = k
	}

}
