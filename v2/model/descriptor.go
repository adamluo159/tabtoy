package model

import "errors"

type DescriptorKind int

const (
	DescriptorKind_None DescriptorKind = iota
	DescriptorKind_Enum
	DescriptorKind_Struct
)

type DescriptorUsage int

const (
	DescriptorUsage_None          DescriptorUsage = iota
	DescriptorUsage_RowType                       // 每个表的行类型
	DescriptorUsage_CombineStruct                 // 最终使用的合并结构体
)

type Descriptor struct {
	Name     string
	Kind     DescriptorKind
	Usage    DescriptorUsage
	NotPrint bool

	FieldByName   map[string]*FieldDescriptor
	FieldByNumber map[int32]*FieldDescriptor
	Fields        []*FieldDescriptor

	Indexes     []*FieldDescriptor
	IndexByName map[string]*FieldDescriptor

	// 联合索引映射，key是联合索引名称，value是该联合索引包含的字段列表
	UnionIndexes map[string][]*FieldDescriptor

	File *FileDescriptor
}

var (
	ErrDuplicateFieldName = errors.New("Duplicate field name")
	ErrDuplicateIndexName = errors.New("Duplicate index name")
)

func (self *Descriptor) Add(def *FieldDescriptor) error {

	// 创建字段
	if _, ok := self.FieldByName[def.Name]; ok {
		return ErrDuplicateFieldName
	} else {
		self.FieldByName[def.Name] = def
		self.FieldByNumber[def.EnumValue] = def
		self.Fields = append(self.Fields, def)
	}

	// 创建单字段索引
	if def.Meta.GetBool("MakeIndex") {

		if _, ok := self.IndexByName[def.Name]; ok {
			return ErrDuplicateIndexName
		} else {
			self.IndexByName[def.Name] = def
			self.Indexes = append(self.Indexes, def)
		}
	}

	// 处理联合索引
	if unionIndexName := def.Meta.GetString("UnionIndex"); unionIndexName != "" {
		// 将字段添加到对应的联合索引列表
		self.UnionIndexes[unionIndexName] = append(self.UnionIndexes[unionIndexName], def)
	}

	return nil
}

// 根据MakeIndex:true+RepeatCheck:true自动生成联合索引
func (self *Descriptor) AutoGenerateUnionIndexes() {
	// 收集所有带有MakeIndex:true+RepeatCheck:true的字段
	var makeIndexFields []*FieldDescriptor
	for _, field := range self.Fields {
		if field.Meta.GetBool("MakeIndex") && field.Meta.GetBool("RepeatCheck") {
			makeIndexFields = append(makeIndexFields, field)
		}
	}

	// 如果有2个或以上这样的字段，自动生成联合索引
	if len(makeIndexFields) >= 2 {
		// 使用第一个字段名作为联合索引名称
		unionIndexName := makeIndexFields[0].Name
		// 将所有这些字段添加到同一个联合索引中
		for _, field := range makeIndexFields {
			// 检查字段是否已经在其他联合索引中
			alreadyInUnionIndex := false
			for _, unionFields := range self.UnionIndexes {
				for _, unionField := range unionFields {
					if unionField == field {
						alreadyInUnionIndex = true
						break
					}
				}
				if alreadyInUnionIndex {
					break
				}
			}
			if !alreadyInUnionIndex {
				self.UnionIndexes[unionIndexName] = append(self.UnionIndexes[unionIndexName], field)
			}
		}
	}
}

func (self *Descriptor) FieldByValueAndMeta(value string) *FieldDescriptor {

	for _, v := range self.FieldByName {

		if v.Name == value {
			return v
		}

		if v.Meta.GetString("Alias") == value {
			return v
		}

	}

	return nil
}

func NewDescriptor() *Descriptor {
	return &Descriptor{
		FieldByName:    make(map[string]*FieldDescriptor),
		FieldByNumber:  make(map[int32]*FieldDescriptor),
		IndexByName:    make(map[string]*FieldDescriptor),
		UnionIndexes:   make(map[string][]*FieldDescriptor),
	}
}
