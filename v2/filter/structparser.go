package filter

import (
	"strings"

	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
	"github.com/davyxu/golexer"
)

// 自定义的token id
const (
	Token_EOF = iota
	Token_WhiteSpace
	Token_LineEnd
	Token_UnixStyleComment
	Token_Identifier
	Token_Numeral
	Token_String
	Token_Comma
	Token_Unknown
)

type structParser struct {
	*golexer.Parser
}

func (self *structParser) Run(fd *model.FieldDescriptor, callback func(string, string) bool) (ok bool) {

	defer golexer.ErrorCatcher(func(err error) {

		log.Errorf("%s, '%s' '%v'", i18n.String(i18n.StructParser_LexerError), fd.Name, err.Error())
	})

	self.NextToken()

	for self.TokenID() != Token_EOF {

		if self.TokenID() != Token_Identifier {
			log.Errorf("%s, '%s'", i18n.String(i18n.StructParser_ExpectField), fd.Name)
			return false
		}

		key := self.TokenValue()

		self.NextToken()

		if self.TokenID() != Token_Comma {
			log.Errorf("%s, '%s'", i18n.String(i18n.StructParser_UnexpectedSpliter), key)
			return false
		}

		self.NextToken()

		value := self.TokenValue()

		if !callback(key, value) {
			return false
		}

		self.NextToken()

	}

	return true
}

func newStructParser(value string) *structParser {
	l := golexer.NewLexer()

	l.AddMatcher(golexer.NewNumeralMatcher(Token_Numeral))
	l.AddMatcher(golexer.NewStringMatcher(Token_String))

	l.AddIgnoreMatcher(golexer.NewWhiteSpaceMatcher(Token_WhiteSpace))
	l.AddIgnoreMatcher(golexer.NewLineEndMatcher(Token_LineEnd))
	l.AddIgnoreMatcher(golexer.NewUnixStyleCommentMatcher(Token_UnixStyleComment))

	l.AddMatcher(golexer.NewSignMatcher(Token_Comma, ":"))

	l.AddMatcher(golexer.NewIdentifierMatcher(Token_Identifier))

	l.AddMatcher(golexer.NewUnknownMatcher(Token_Unknown))

	l.Start(value)

	return &structParser{
		golexer.NewParser(l, value),
	}

}

func parseStruct(fd *model.FieldDescriptor, value string, fileD *model.FileDescriptor, node *model.Node) bool {
	// 检查是否支持简化输入：结构体有2个字段，且第一个字段标记了 SimpleInput:true
	if len(fd.Complex.Fields) == 2 && fd.Complex.Fields[0].Meta.GetBool("SimpleInput") {
		if result := parseSimpleStruct(fd, value, fileD, node); result {
			return true
		}
		// 简化格式解析失败，继续尝试标准格式
	}

	p := newStructParser(value)

	// 检查字段有没有重复
	sfList := newStructFieldList()

	result := p.Run(fd, func(key, value string) bool {

		bnField := fd.Complex.FieldByValueAndMeta(key)
		if bnField == nil {

			log.Errorf("%s, '%s'", i18n.String(i18n.StructParser_FieldNotFound), key)

			return false
		}

		if sfList.Exists(bnField) {
			log.Errorf("%s, '%s'", i18n.String(i18n.StructParser_DuplicateFieldInCell), key)
			return false
		}

		sfList.Add(bnField, value)

		return true
	})

	if !result {
		return false
	}

	// 结构体中未填的字段如果是Default, 也要输出
	for _, structField := range fd.Complex.Fields {

		if sfList.Exists(structField) {
			continue
		}

		if structField.Meta.GetString("Default") != "" {
			sfList.Add(structField, structField.Meta.GetString("Default"))
		}

	}

	// 结构体输出是map顺序, 必须按照定义时的order进行排序, 否则在二进制中顺序是错的
	sfList.Sort()

	for i := 0; i < sfList.Len(); i++ {

		v := sfList.Get(i)

		// 添加类型节点
		fieldNode := node.AddKey(v.key)

		// 在类型节点下添加值节点
		_, ok := ConvertValue(v.key, v.value, fileD, fieldNode)

		if !ok {
			return false
		}

	}

	return true

}

// parseSimpleStruct 解析简化格式的结构体输入
// 格式: value1:value2，例如 "金币:2000" 或 "10000:2000"
// 注意：简化格式不能包含空格，否则会被认为是标准格式
func parseSimpleStruct(fd *model.FieldDescriptor, value string, fileD *model.FileDescriptor, node *model.Node) bool {
	// 检查是否包含空格，如果包含空格则不是简化格式
	if strings.Contains(value, " ") {
		return false
	}

	// 查找冒号位置
	colonIndex := findColonIndex(value)
	if colonIndex == -1 {
		return false
	}

	field1Value := trimString(value[:colonIndex])
	field2Value := trimString(value[colonIndex+1:])

	// 获取结构体的两个字段
	field1 := fd.Complex.Fields[0]
	field2 := fd.Complex.Fields[1]

	// 解析第一个字段
	fieldNode1 := node.AddKey(field1)
	if _, ok := ConvertValue(field1, field1Value, fileD, fieldNode1); !ok {
		return false
	}

	// 解析第二个字段
	fieldNode2 := node.AddKey(field2)
	if _, ok := ConvertValue(field2, field2Value, fileD, fieldNode2); !ok {
		return false
	}

	return true
}
