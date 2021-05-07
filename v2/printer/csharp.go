package printer

import (
	"fmt"
	"text/template"

	"bytes"
	"crypto/md5"
	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
	"strings"
)

const csharpTemplate = `// Generated by github.com/adamluo159/tabtoy
// Version: {{.ToolVersion}}
// DO NOT EDIT!!
using System.Collections.Generic;

namespace {{.Namespace}}{{$globalIndex:=.Indexes}}{{$verticalFields:=.VerticalFields}}
{
	{{range .Enums}}
	// Defined in table: {{.DefinedTable}}
	public enum {{.Name}}
	{
	{{range .Fields}}	
		{{.Comment}}
		{{.FieldDescriptor.Name}} = {{.FieldDescriptor.EnumValue}}, {{.Alias}}
	{{end}}
	}
	{{end}}
	{{range .Classes}}

	// Defined in table: {{.DefinedTable}}
	{{.CSClassHeader}}
	public partial class {{.Name}}
	{
	{{if .IsCombine}}
		public tabtoy.Logger TableLogger = new tabtoy.Logger();
	{{end}}
	{{range .Fields}}	
		{{.Comment}}
		{{.TypeCode}} {{.Alias}}
	{{end}}
	{{if .IsCombine}}
		#region Index code
	 	{{range $globalIndex}}Dictionary<{{.IndexType}}, {{.RowType}}> _{{.RowName}}By{{.IndexName}} = new Dictionary<{{.IndexType}}, {{.RowType}}>();
        public {{.RowType}} Get{{.RowName}}By{{.IndexName}}({{.IndexType}} {{.IndexName}}, {{.RowType}} def = default({{.RowType}}))
        {
            {{.RowType}} ret;
            if ( _{{.RowName}}By{{.IndexName}}.TryGetValue( {{.IndexName}}, out ret ) )
            {
                return ret;
            }
			
			if ( def == default({{.RowType}}) )
			{
				TableLogger.ErrorLine("Get{{.RowName}}By{{.IndexName}} failed, {{.IndexName}}: {0}", {{.IndexName}});
			}

            return def;
        }
		{{end}}
		public string GetBuildID(){
			return "{{$.BuildID}}";
		}
	{{range $verticalFields}}
		public {{.StructName}} Get{{.Name}}( )
		{
			return {{.Name}}[0];
		}	
	{{end}}
		#endregion
		#region Deserialize code
		{{range $.Classes}}
		static tabtoy.DeserializeHandler<{{.Name}}> _{{.Name}}DeserializeHandler;
		static tabtoy.DeserializeHandler<{{.Name}}> {{.Name}}DeserializeHandler
		{
			get
			{
				if (_{{.Name}}DeserializeHandler == null )
				{
					_{{.Name}}DeserializeHandler = new tabtoy.DeserializeHandler<{{.Name}}>(Deserialize);
				}

				return _{{.Name}}DeserializeHandler;
			}
		}
		public static void Deserialize( {{.Name}} ins, tabtoy.DataReader reader )
		{
			{{ if $.GenSerializeCode }}
 			int tag = -1;
            while ( -1 != (tag = reader.ReadTag()))
            {
                switch (tag)
                { {{range .Fields}}
                	case {{.Tag}}:
                	{
						{{.ReadCode}}
                	}
                	break; {{end}}
                }
             } {{end}}

			{{range $a, $row :=.IndexedFields}}
			// Build {{$row.FieldDescriptor.Name}} Index
			for( int i = 0;i< ins.{{$row.FieldDescriptor.Name}}.Count;i++)
			{
				var element = ins.{{$row.FieldDescriptor.Name}}[i];
				{{range $b, $key := .IndexKeys}}
				ins._{{$row.FieldDescriptor.Name}}By{{$key.Name}}.Add(element.{{$key.Name}}, element);
				{{end}}
			}
			{{end}}
		}{{end}}
		#endregion
		#region Clear Code
		public void Clear( )
		{	{{range .Fields}}		
				{{.Name}}.Clear(); {{end}}
			{{range $globalIndex}}
				_{{.RowName}}By{{.IndexName}}.Clear(); {{end}}
		}
		#endregion
	{{end}}

	} {{end}}

}
`

type indexField struct {
	TableIndex
}

func (self indexField) IndexName() string {
	return self.Index.Name
}

func (self indexField) RowType() string {
	return self.Row.Complex.Name
}

func (self indexField) RowName() string {
	return self.Row.Name
}

func (self indexField) IndexType() string {

	switch self.Index.Type {
	case model.FieldType_Int32:
		return "int"
	case model.FieldType_UInt32:
		return "uint"
	case model.FieldType_Int64:
		return "long"
	case model.FieldType_UInt64:
		return "ulong"
	case model.FieldType_String:
		return "string"
	case model.FieldType_Float:
		return "float"
	case model.FieldType_Bool:
		return "bool"
	case model.FieldType_Enum:

		return self.Index.Complex.Name
	default:
		log.Errorf("%s can not be index ", self.Index.String())
	}

	return "unknown"
}

type csharpField struct {
	*model.FieldDescriptor

	IndexKeys []*model.FieldDescriptor

	parentStruct *structModel
}

func (self csharpField) Alias() string {

	v := self.FieldDescriptor.Meta.GetString("Alias")
	if v == "" {
		return ""
	}

	return "// " + v
}

func (self csharpField) Comment() string {

	if self.FieldDescriptor.Comment == "" {
		return ""
	}

	// zjwps 建议修改
	return "/// <summary> \n		/// " + strings.Replace(self.FieldDescriptor.Comment, "\n", "\n		///", -1) + "\n		/// </summary>"
}

func (self csharpField) ReadCode() string {

	var baseType string

	var descHandlerCode string

	switch self.Type {
	case model.FieldType_Int32:
		baseType = "Int32"
	case model.FieldType_UInt32:
		baseType = "UInt32"
	case model.FieldType_Int64:
		baseType = "Int64"
	case model.FieldType_UInt64:
		baseType = "UInt64"
	case model.FieldType_String:
		baseType = "String"
	case model.FieldType_Float:
		baseType = "Float"
	case model.FieldType_Bool:
		baseType = "Bool"
	case model.FieldType_Enum:

		if self.Complex == nil {
			return "unknown"
		}

		baseType = fmt.Sprintf("Enum<%s>", self.Complex.Name)

	case model.FieldType_Struct:
		if self.Complex == nil {
			return "unknown"
		}

		baseType = fmt.Sprintf("Struct<%s>", self.Complex.Name)

	}

	if self.Type == model.FieldType_Struct {
		descHandlerCode = fmt.Sprintf("%sDeserializeHandler", self.Complex.Name)
	}

	if self.IsRepeated {
		return fmt.Sprintf("ins.%s.Add( reader.Read%s(%s) );", self.Name, baseType, descHandlerCode)
	} else {
		return fmt.Sprintf("ins.%s = reader.Read%s(%s);", self.Name, baseType, descHandlerCode)
	}

}

func (self csharpField) Tag() string {

	if self.parentStruct.IsCombine() {
		tag := model.MakeTag(int32(model.FieldType_Table), self.Order)

		return fmt.Sprintf("0x%x", tag)
	}

	return fmt.Sprintf("0x%x", self.FieldDescriptor.Tag())
}

func (self csharpField) StructName() string {
	if self.Complex == nil {
		return "[NotComplex]"
	}

	return self.Complex.Name
}

func (self csharpField) IsVerticalStruct() bool {
	if self.FieldDescriptor.Complex == nil {
		return false
	}

	return self.FieldDescriptor.Complex.File.Pragma.GetBool("Vertical")
}

func (self csharpField) TypeCode() string {

	var raw string

	switch self.Type {
	case model.FieldType_Int32:
		raw = "int"
	case model.FieldType_UInt32:
		raw = "uint"
	case model.FieldType_Int64:
		raw = "long"
	case model.FieldType_UInt64:
		raw = "ulong"
	case model.FieldType_String:
		raw = "string"
	case model.FieldType_Float:
		raw = "float"
	case model.FieldType_Bool:
		raw = "bool"
	case model.FieldType_Enum:
		if self.Complex == nil {
			log.Errorln("unknown enum type ", self.Type)
			return "unknown"
		}

		raw = self.Complex.Name
	case model.FieldType_Struct:
		if self.Complex == nil {
			log.Errorln("unknown struct type ", self.Type, self.FieldDescriptor.Name, self.FieldDescriptor.Parent.Name)
			return "unknown"
		}

		raw = self.Complex.Name

		// 非repeated的结构体
		if !self.IsRepeated {
			return fmt.Sprintf("public %s %s = new %s();", raw, self.Name, raw)
		}

	default:
		raw = "unknown"
	}

	if self.IsRepeated {
		return fmt.Sprintf("public List<%s> %s = new List<%s>();", raw, self.Name, raw)
	}

	return fmt.Sprintf("public %s %s = %s;", raw, self.Name, wrapCSharpDefaultValue(self.FieldDescriptor))
}

func wrapCSharpDefaultValue(fd *model.FieldDescriptor) string {
	switch fd.Type {
	case model.FieldType_Enum:
		return fmt.Sprintf("%s.%s", fd.Complex.Name, fd.DefaultValue())
	case model.FieldType_String:
		return fmt.Sprintf("\"%s\"", fd.DefaultValue())
	case model.FieldType_Float:
		return fmt.Sprintf("%sf", fd.DefaultValue())
	}

	return fd.DefaultValue()
}

type structModel struct {
	*model.Descriptor
	Fields        []csharpField
	IndexedFields []csharpField // 与csharpField.IndexKeys组成树状的索引层次
}

func (self *structModel) CSClassHeader() string {

	// zjwps 提供需求
	return self.File.Pragma.GetString("CSClassHeader")
}

func (self *structModel) DefinedTable() string {
	return self.File.Name
}

func (self *structModel) Name() string {
	return self.Descriptor.Name
}

func (self *structModel) IsCombine() bool {
	return self.Descriptor.Usage == model.DescriptorUsage_CombineStruct
}

type csharpFileModel struct {
	Namespace   string
	ToolVersion string
	Classes     []*structModel
	Enums       []*structModel
	Indexes     []indexField // 全局的索引

	VerticalFields []csharpField

	GenSerializeCode bool

	BuildID string
}

type csharpPrinter struct {
}

func (self *csharpPrinter) Run(g *Globals) *Stream {

	tpl, err := template.New("csharp").Parse(csharpTemplate)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	var m csharpFileModel

	if g.PackageName != "" {
		m.Namespace = g.PackageName
	} else {
		m.Namespace = g.FileDescriptor.Pragma.GetString("Package")
	}

	m.ToolVersion = g.Version
	m.GenSerializeCode = g.GenCSSerailizeCode

	// combinestruct的全局索引
	for _, ti := range g.GlobalIndexes {

		// 索引也限制
		if !ti.Index.Parent.File.MatchTag(".cs") {
			continue
		}

		m.Indexes = append(m.Indexes, indexField{TableIndex: ti})
	}

	// 遍历所有类型
	for _, d := range g.FileDescriptor.Descriptors {

		// 这给被限制输出
		if !d.File.MatchTag(".cs") {
			log.Infof("%s: %s", i18n.String(i18n.Printer_IgnoredByOutputTag), d.Name)
			continue
		}

		var sm structModel
		sm.Descriptor = d

		switch d.Kind {
		case model.DescriptorKind_Struct:
			m.Classes = append(m.Classes, &sm)
		case model.DescriptorKind_Enum:
			m.Enums = append(m.Enums, &sm)
		}

		// 遍历字段
		for _, fd := range d.Fields {

			// 对CombineStruct的XXDefine对应的字段
			if d.Usage == model.DescriptorUsage_CombineStruct {

				// 这个字段被限制输出
				if fd.Complex != nil && !fd.Complex.File.MatchTag(".cs") {
					continue
				}

				// 这个结构有索引才创建
				if fd.Complex != nil && len(fd.Complex.Indexes) > 0 {

					// 被索引的结构
					indexedField := csharpField{FieldDescriptor: fd, parentStruct: &sm}

					// 索引字段
					for _, key := range fd.Complex.Indexes {
						indexedField.IndexKeys = append(indexedField.IndexKeys, key)
					}

					sm.IndexedFields = append(sm.IndexedFields, indexedField)
				}

				if fd.Complex != nil && fd.Complex.File.Pragma.GetBool("Vertical") {
					m.VerticalFields = append(m.VerticalFields, csharpField{FieldDescriptor: fd, parentStruct: &sm})
				}

			}

			csField := csharpField{FieldDescriptor: fd, parentStruct: &sm}

			sm.Fields = append(sm.Fields, csField)

		}

	}

	bf := NewStream()

	var md5Buffer bytes.Buffer
	err = tpl.Execute(&md5Buffer, &m)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	m.BuildID = fmt.Sprintf("%x", md5.Sum(md5Buffer.Bytes()))
	g.BuildID = m.BuildID

	err = tpl.Execute(bf.Buffer(), &m)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	return bf
}

func init() {

	RegisterPrinter("cs", &csharpPrinter{})

}
