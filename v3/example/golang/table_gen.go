// Generated by github.com/adamluo159/tabtoy
// DO NOT EDIT!!
// Version:
package main

import "errors"

type TableEnumValue struct {
	Name  string
	Index int32
}

type ActorType int32

const (
	ActorType_None    = 0 //
	ActorType_Pharah  = 1 // 法鸡
	ActorType_Junkrat = 2 // 狂鼠
	ActorType_Genji   = 3 // 源氏
	ActorType_Mercy   = 4 // 天使
)

var (
	ActorTypeEnumValues = []TableEnumValue{
		{Name: "None", Index: 0},    //
		{Name: "Pharah", Index: 1},  // 法鸡
		{Name: "Junkrat", Index: 2}, // 狂鼠
		{Name: "Genji", Index: 3},   // 源氏
		{Name: "Mercy", Index: 4},   // 天使
	}
	ActorTypeMapperValueByName = map[string]int32{}
	ActorTypeMapperNameByValue = map[int32]string{}
)

func (self ActorType) String() string {
	name, _ := ActorTypeMapperNameByValue[int32(self)]
	return name
}

type ExampleData struct {
	ID       int32     `tb_name:"任务ID"`
	ID2      int32     `tb_name:"任务ID2"`
	Name     string    `tb_name:"名称"`
	Rate     float32   `tb_name:"比例"`
	Accuracy float64   `tb_name:"精度"`
	Type     ActorType `tb_name:"类型"`
	Skill    []int32   `tb_name:"技能列表"`
	Buff     int32     `tb_name:"增益"`
	TagList  []string  `tb_name:"标记"`
	Multi    []int32   `tb_name:"多列"`
}

type ExtendData struct {
	Additive float32 `tb_name:"附加"`
	Index2   int32   `tb_name:"索引2"`
}

type ExampleKV struct {
	ServerIP   string  `tb_name:"服务器IP"`
	ServerPort uint16  `tb_name:"服务器端口"`
	GroupID    []int32 `tb_name:"分组"`
}

// Combine struct
type Table struct {
	ExampleData []*ExampleData // table: ExampleData
	ExtendData  []*ExtendData  // table: ExtendData
	ExampleKV   []*ExampleKV   // table: ExampleKV

	// Indices
	ExampleDataByID    map[int32]*ExampleData `json:"-"` // table: ExampleData
	ExampleDataByID2   map[int32]*ExampleData `json:"-"` // table: ExampleData
	ExtendDataByIndex2 map[int32]*ExtendData  `json:"-"` // table: ExtendData

	// Handlers
	postHandlers []func(*Table) error `json:"-"`
	preHandlers  []func(*Table) error `json:"-"`

	indexHandler map[string]func() `json:"-"`
	resetHandler map[string]func() `json:"-"`
}

// table: ExampleKV
func (self *Table) GetKeyValue_ExampleKV() *ExampleKV {
	return self.ExampleKV[0]
}

// 注册加载后回调(用于构建数据)
func (self *Table) RegisterPostEntry(h func(*Table) error) {

	if h == nil {
		panic("empty postload handler")
	}

	self.postHandlers = append(self.postHandlers, h)
}

// 注册加载前回调(用于清除数据)
func (self *Table) RegisterPreEntry(h func(*Table) error) {

	if h == nil {
		panic("empty preload handler")
	}

	self.preHandlers = append(self.preHandlers, h)
}

// 清除索引和数据
func (self *Table) ResetData() error {

	err := self.InvokePreHandler()
	if err != nil {
		return err
	}

	return self.ResetTable("")
}

// 全局表构建索引及通知回调
func (self *Table) BuildData() error {

	err := self.IndexTable("")
	if err != nil {
		return err
	}

	return self.InvokePostHandler()
}

// 调用加载前回调
func (self *Table) InvokePreHandler() error {
	for _, h := range self.preHandlers {
		if err := h(self); err != nil {
			return err
		}
	}

	return nil
}

// 调用加载后回调
func (self *Table) InvokePostHandler() error {
	for _, h := range self.postHandlers {
		if err := h(self); err != nil {
			return err
		}
	}

	return nil
}

// 为表建立索引. 表名为空时, 构建所有表索引
func (self *Table) IndexTable(tableName string) error {

	if tableName == "" {

		for _, h := range self.indexHandler {
			h()
		}
		return nil

	} else {
		if h, ok := self.indexHandler[tableName]; ok {
			h()
		}

		return nil
	}
}

// 重置表格数据
func (self *Table) ResetTable(tableName string) error {
	if tableName == "" {
		for _, h := range self.resetHandler {
			h()
		}

		return nil
	} else {
		if h, ok := self.resetHandler[tableName]; ok {
			h()
			return nil
		}

		return errors.New("reset table failed, table not found: " + tableName)
	}
}

// 初始化表实例
func NewTable() *Table {

	self := &Table{
		indexHandler: make(map[string]func()),
		resetHandler: make(map[string]func()),
	}

	self.indexHandler["ExampleData"] = func() {
		for _, v := range self.ExampleData {
			self.ExampleDataByID[v.ID] = v
		}
	}

	self.indexHandler["ExampleData"] = func() {
		for _, v := range self.ExampleData {
			self.ExampleDataByID2[v.ID2] = v
		}
	}

	self.indexHandler["ExtendData"] = func() {
		for _, v := range self.ExtendData {
			self.ExtendDataByIndex2[v.Index2] = v
		}
	}

	self.resetHandler["ExampleData"] = func() {
		self.ExampleData = nil

		self.ExampleDataByID = map[int32]*ExampleData{}
		self.ExampleDataByID2 = map[int32]*ExampleData{}
	}
	self.resetHandler["ExtendData"] = func() {
		self.ExtendData = nil

		self.ExtendDataByIndex2 = map[int32]*ExtendData{}
	}
	self.resetHandler["ExampleKV"] = func() {
		self.ExampleKV = nil

	}

	self.ResetData()

	return self
}

func init() {

	for _, v := range ActorTypeEnumValues {
		ActorTypeMapperValueByName[v.Name] = v.Index
		ActorTypeMapperNameByValue[v.Index] = v.Name
	}

}
