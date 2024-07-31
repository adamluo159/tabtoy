// Generated by github.com/adamluo159/tabtoy
// Version:
// DO NOT EDIT!!
package table

import (
	"fmt"
	"encoding/json"
	"os"
)

// Defined in table: Globals
type ActorType int32

const (

	//唐僧
	ActorType_Leader ActorType = 0

	//孙悟空
	ActorType_Monkey ActorType = 1

	//猪八戒
	ActorType_Pig ActorType = 2

	//沙僧
	ActorType_Hammer ActorType = 3
)

var (
	ActorTypeMapperValueByName = map[string]int32{
		"Leader": 0,
		"Monkey": 1,
		"Pig":    2,
		"Hammer": 3,
	}

	ActorTypeMapperNameByValue = map[int32]string{
		0: "Leader",
		1: "Monkey",
		2: "Pig",
		3: "Hammer",
	}

	ActorTypeCommentByValue = map[int32]string{
		0: "唐僧",
		1: "孙悟空",
		2: "猪八戒",
		3: "沙僧",
	}
)

func (self ActorType) String() string {
	name, _ := ActorTypeMapperNameByValue[int32(self)]
	return name
}

// Defined in table: Config
type Config struct {

	//AAA
	AAA []*AAADefine
}

// Defined in table: Globals
type Vec2 struct {
	X int32

	Y int32
}

// Defined in table: AAA
type AAADefine struct {

	//唯一ID
	ID int32

	//名称
	Name string

	//名称
	SSS string `Mark:"Client"`

	DDD *Vec2
}

// Config 访问接口
type ConfigTable struct {

	// 表格原始数据
	Config

	// 索引函数表
	indexFuncByName map[string][]func(*ConfigTable) error

	// 清空函数表
	clearFuncByName map[string][]func(*ConfigTable) error

	// 加载前回调
	preFuncList []func(*ConfigTable) error

	// 加载后回调
	postFuncList []func(*ConfigTable) error

	AAAByID map[int32]*AAADefine
}

// 从json文件加载
func (self *ConfigTable) Load(filename string) error {
	data, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	return self.LoadData(data)
}

// 从二进制加载
func (self *ConfigTable) LoadData(data []byte) error {

	var newTab Config

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
	self.Config = newTab

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
func (self *ConfigTable) RegisterIndexEntry(name string, indexCallback func(*ConfigTable) error, clearCallback func(*ConfigTable) error) {

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
func (self *ConfigTable) RegisterPreEntry(callback func(*ConfigTable) error) {

	self.preFuncList = append(self.preFuncList, callback)
}

// 注册所有完成时回调
func (self *ConfigTable) RegisterPostEntry(callback func(*ConfigTable) error) {

	self.postFuncList = append(self.postFuncList, callback)
}

// 创建一个Config表读取实例
func NewConfigTable() *ConfigTable {
	return &ConfigTable{

		indexFuncByName: map[string][]func(*ConfigTable) error{

			"AAA": {func(tab *ConfigTable) error {

				// AAA
				for _, def := range tab.AAA {

					if _, ok := tab.AAAByID[def.ID]; ok {
						panic(fmt.Sprintf("duplicate index in AAAByID: %v", def.ID))
					}

					tab.AAAByID[def.ID] = def

				}

				return nil
			}},
		},

		clearFuncByName: map[string][]func(*ConfigTable) error{

			"AAA": {func(tab *ConfigTable) error {

				// AAA

				tab.AAAByID = make(map[int32]*AAADefine)

				return nil
			}},
		},

		AAAByID: make(map[int32]*AAADefine),
	}
}
