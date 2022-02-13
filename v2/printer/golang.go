package printer

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"text/template"

	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
)

const goTemplate = `// Generated by github.com/adamluo159/tabtoy
// Version: {{.ToolVersion}}
// DO NOT EDIT!!
package {{.Package}}

import(	
	{{if .HasAnyIndex}}"fmt"{{end}}
	"encoding/json"
	"io/ioutil"
)
{{range $a, $en := .Enums}} 
// Defined in table: {{$en.DefinedTable}}
type {{$en.Name}} int32
const (	
{{range .GoFields}}
	{{.Comment}}
	{{$en.Name}}_{{.Name}} {{$en.Name}} = {{.Number}}
{{end}}
)
var (
{{$en.Name}}MapperValueByName = map[string]int32{ {{range .GoFields}}
	"{{.Name}}": {{.Number}}, {{end}}
}

{{$en.Name}}MapperNameByValue = map[int32]string{ {{range .GoFields}}
	{{.Number}}: "{{.Name}}" , {{end}}
}

{{$en.Name}}CommentByValue = map[int32]string{ {{range .GoFields}}
	{{.Number}}: "{{.RawComment}}" , {{end}}
}

)

func (self {{$en.Name}}) String() string {
	name, _ := {{$en.Name}}MapperNameByValue[int32(self)]
	return name
}
{{end}}

{{range $a, $strus := .Structs}} 
// Defined in table: {{$strus.DefinedTable}}
type {{$strus.Name}} struct{
	{{range $b, $fd := $strus.GoFields}} 
	{{.Comment}}
	{{$fd.Name}} {{$fd.TypeString}} {{$fd.StructTag}}
	{{end}}
}
{{end}}

// {{$.Name}} 访问接口
type {{$.Name}}Table struct{
	
	// 表格原始数据
	{{$.Name}}
	
	// 索引函数表
	indexFuncByName map[string][]func(*{{$.Name}}Table) error

	// 清空函数表
	clearFuncByName map[string][]func(*{{$.Name}}Table) error

	// 加载前回调
	preFuncList []func(*{{$.Name}}Table) error

	// 加载后回调
	postFuncList []func(*{{$.Name}}Table) error
	
	{{range $a, $strus := .IndexedStructs}} {{range .Indexes}}
	{{$strus.Name}}By{{.Name}} map[{{.KeyType}}]*{{$strus.TypeName}}
	{{end}} {{end}}
}

{{range .VerticalFields}}
{{.Comment}}
func (self *{{$.Name}}Table) Get{{.Name}}( ) {{.ElementTypeString}} {
	return self.{{.Name}}[0]
}
{{end}}


// 从json文件加载
func (self *{{$.Name}}Table) Load(filename string) error {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	return self.LoadData(data)
}

// 从二进制加载
func (self *{{$.Name}}Table) LoadData(data []byte) error {

	var newTab {{$.Name}}

	// 读取
	err := json.Unmarshal(data, &newTab)
	if err != nil {
		return err
	}

	// 所有加载前的回调
	for _, v := range self.preFuncList {
		if err = v(self); err != nil {
			return err
		}
	}
	
	// 清除前通知
	for _, list := range self.clearFuncByName {
		for _, v := range list {
			if err = v(self); err != nil {
				return err
			}
		}
	}

	// 复制数据
	self.{{$.Name}} = newTab

	// 生成索引
	for _, list := range self.indexFuncByName {
		for _, v := range list {
			if err = v(self); err != nil {
				return err
			}
		}
	}

	// 所有完成时的回调
	for _, v := range self.postFuncList {
		if err = v(self); err != nil {
			return err
		}
	}

	return nil
}

// 注册外部索引入口, 索引回调, 清空回调
func (self *{{$.Name}}Table) RegisterIndexEntry(name string, indexCallback func(*{{$.Name}}Table) error, clearCallback func(*{{$.Name}}Table)error) {

	indexList, _ := self.indexFuncByName[name]
	clearList, _ := self.clearFuncByName[name]

	if indexCallback != nil {
		indexList = append(indexList, indexCallback)
	}

	if clearCallback != nil {
		clearList = append(clearList, clearCallback)
	}

	self.indexFuncByName[name] = indexList
	self.clearFuncByName[name] = clearList
}

// 注册加载前回调
func (self *{{$.Name}}Table) RegisterPreEntry(callback func(*{{$.Name}}Table) error) {

	self.preFuncList = append(self.preFuncList, callback)
}


// 注册所有完成时回调
func (self *{{$.Name}}Table) RegisterPostEntry(callback func(*{{$.Name}}Table) error) {

	self.postFuncList = append(self.postFuncList, callback)
}


// 创建一个{{$.Name}}表读取实例
func New{{$.Name}}Table() *{{$.Name}}Table {
	return &{{$.Name}}Table{

	
		indexFuncByName: map[string][]func(*{{$.Name}}Table) error{
		
		{{range $a, $strus := .IndexedStructs}}
			"{{$strus.Name}}": {func(tab *{{$.Name}}Table)error {
				
				// {{$strus.Name}}
				for _, def := range tab.{{$strus.Name}} {
					{{range .Indexes}}
					if _, ok := tab.{{$strus.Name}}By{{.Name}}[def.{{.Name}}]; ok {
						panic(fmt.Sprintf("duplicate index in {{$strus.Name}}By{{.Name}}: %v", def.{{.Name}}))
					}
					{{end}}		
					{{range .Indexes}}
					tab.{{$strus.Name}}By{{.Name}}[def.{{.Name}}] = def{{end}}
					
				}

				return nil
			}},
		{{end}}
		
			
		},
		
		clearFuncByName: map[string][]func(*{{$.Name}}Table)error{
		
		{{range $a, $strus := .IndexedStructs}}
			"{{$strus.Name}}": {func(tab *{{$.Name}}Table) error{
				
				// {{$strus.Name}}
	
				{{range .Indexes}}
				tab.{{$strus.Name}}By{{.Name}} = make(map[{{.KeyType}}]*{{$strus.TypeName}}){{end}}

				return nil
			}},
		{{end}}
		
			
		},
		

		{{range $a, $strus := .IndexedStructs}} {{range .Indexes}}
		{{$strus.Name}}By{{.Name}} : make(map[{{.KeyType}}]*{{$strus.TypeName}}),
		{{end}} {{end}}
		
	}
}


`

// 每个带有MakeIndex的列
type goFieldModel struct {
	*model.FieldDescriptor
	Number int
}

func (self *goFieldModel) Alias() string {
	return self.FieldDescriptor.Meta.GetString("Alias")
}

func (self *goFieldModel) RawComment() string {
	var out string

	if self.FieldDescriptor.Meta.GetString("Alias") != "" {
		return self.FieldDescriptor.Meta.GetString("Alias")
	}

	if self.FieldDescriptor.Comment != "" {
		return self.FieldDescriptor.Comment
	}
	return out
}

func (self *goFieldModel) Comment() string {

	var out string

	if self.FieldDescriptor.Meta.GetString("Alias") != "" {
		out += "// "
		out += self.FieldDescriptor.Meta.GetString("Alias")
	}

	if self.FieldDescriptor.Comment != "" {
		if out == "" {
			out += "//"
		}
		out += self.FieldDescriptor.Comment
	}

	return out

}

func (self *goFieldModel) KeyType() string {

	// 修复: 当枚举做索引时, 多出包名
	if self.Type == model.FieldType_Enum {
		return self.Complex.Name
	}

	return model.FieldTypeToString(self.Type)
}

func (self *goFieldModel) StructTag() string {

	var buf bytes.Buffer

	buf.WriteString("`")

	var userTagCount int

	self.Meta.VisitUserMeta(func(k string, v interface{}) bool {

		if userTagCount > 0 {
			buf.WriteString(" ")
		}

		buf.WriteString(fmt.Sprintf("%s:\"%s\"", k, v))

		userTagCount++

		return true
	})

	buf.WriteString("`")

	if userTagCount == 0 {
		return ""
	}

	return buf.String()
}

func (self *goFieldModel) ElementTypeString() string {

	switch self.FieldDescriptor.Type {
	case model.FieldType_Float:
		return "float32"
	case model.FieldType_Struct:
		return "*" + self.FieldDescriptor.TypeString()
	default:
		return self.FieldDescriptor.TypeString()
	}
}

func (self *goFieldModel) TypeString() string {

	var prefix string
	if self.IsRepeated {
		prefix = "[]"
	}

	return prefix + self.ElementTypeString()

}

// 对应每个电子表格文件
type goIndexStructModel struct {
	*model.FieldDescriptor

	Indexes []*goFieldModel
}

func (self *goIndexStructModel) TypeName() string {
	return self.Complex.Name
}

type goStructModel struct {
	*model.Descriptor

	GoFields []*goFieldModel
}

func (self *goStructModel) DefinedTable() string {
	return self.File.Name
}

// 整个导出文件
type goFileModel struct {
	*model.FileDescriptor
	ToolVersion    string
	IndexedStructs []*goIndexStructModel
	Structs        []*goStructModel
	Enums          []*goStructModel
	IndexCount     int

	// 配置的字段
	VerticalFields []*goFieldModel
}

func (self *goFileModel) HasAnyIndex() bool {
	return self.IndexCount > 0
}

func (self *goFileModel) HasAnyStruct() bool {
	return len(self.Structs) > 0
}

func (self *goFileModel) Package() string {
	return self.FileDescriptor.Pragma.GetString("Package")
}

type goPrinter struct {
}

func collectIndexInfo(g *Globals, fm *goFileModel) {

	// 遍历需要导出的合并字段
	for _, fd := range g.CombineStruct.Fields {

		// fd --> 对应每个文件的Row定义, 也就是XXDefine, 在CombineStruct上, 只是一个字段

		// 对CombineStruct的XXDefine对应的字段
		if g.CombineStruct.Usage != model.DescriptorUsage_CombineStruct {
			continue
		}

		//非结构体不输出(其实根本不会击中)
		if fd.Complex == nil {
			continue
		}

		// 这个字段被限制输出
		if !fd.Complex.File.MatchTag(".go") {
			continue
		}

		// 这个结构有索引才创建
		if len(fd.Complex.Indexes) == 0 {
			continue
		}

		rm := goIndexStructModel{FieldDescriptor: fd}

		// 索引字段
		for _, key := range fd.Complex.Indexes {

			rm.Indexes = append(rm.Indexes, &goFieldModel{
				FieldDescriptor: key,
			})
			fm.IndexCount++
		}

		fm.IndexedStructs = append(fm.IndexedStructs, &rm)

	}
}

func collectAllStructInfo(g *Globals, fm *goFileModel) {

	// 遍历所有类型
	for _, d := range g.FileDescriptor.Descriptors {

		// 这给被限制输出
		if !d.File.MatchTag(".go") {
			log.Infof("%s: %s", i18n.String(i18n.Printer_IgnoredByOutputTag), d.Name)
			continue
		}

		structM := &goStructModel{Descriptor: d}

		// 遍历字段
		for index, fd := range d.Fields {

			// 对CombineStruct的XXDefine对应的字段
			if d.Usage == model.DescriptorUsage_CombineStruct {

				// 这个字段被限制输出
				if fd.Complex != nil && !fd.Complex.File.MatchTag(".go") {
					continue
				}

				if fd.Complex != nil && fd.Complex.File.Pragma.GetBool("Vertical") {
					fm.VerticalFields = append(fm.VerticalFields, &goFieldModel{FieldDescriptor: fd})
				}
			}

			field := &goFieldModel{FieldDescriptor: fd}

			switch d.Kind {
			case model.DescriptorKind_Struct:
				field.Number = index + 1
			case model.DescriptorKind_Enum:
				field.Number = int(fd.EnumValue)
			}

			structM.GoFields = append(structM.GoFields, field)

		}

		switch d.Kind {
		case model.DescriptorKind_Struct:
			fm.Structs = append(fm.Structs, structM)
		case model.DescriptorKind_Enum:
			fm.Enums = append(fm.Enums, structM)
		}

	}
}

func (self *goPrinter) Run(g *Globals) *Stream {

	tpl, err := template.New("golang").Parse(goTemplate)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	var fm goFileModel
	fm.ToolVersion = g.Version
	fm.FileDescriptor = g.FileDescriptor

	collectIndexInfo(g, &fm)
	collectAllStructInfo(g, &fm)

	bf := NewStream()

	err = tpl.Execute(bf.Buffer(), &fm)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	if err := formatCode(bf.Buffer()); err != nil {
		log.Errorln("format golang code err", err)
		log.Errorln(bf.Buffer().String())
		return nil
	}

	return bf
}

func formatCode(bf *bytes.Buffer) error {

	fset := token.NewFileSet()

	ast, err := parser.ParseFile(fset, "", bf, parser.ParseComments)
	if err != nil {
		return err
	}

	bf.Reset()

	err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(bf, fset, ast)
	if err != nil {
		return err
	}

	return nil
}

func init() {

	RegisterPrinter("go", &goPrinter{})

}
