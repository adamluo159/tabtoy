package v2

import (
	"path/filepath"
	"strings"

	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
	"github.com/adamluo159/tabtoy/v2/printer"
)

func solveMapFieldTypes(file *File, globalFD *model.FileDescriptor) bool {
	for _, fd := range file.LocalFD.Descriptors {
		for _, field := range fd.Fields {
			if field.Type == model.FieldType_Map && field.RawFieldType != "" {
				if !solveMapFieldType(field, file.LocalFD, globalFD) {
					return false
				}
			}
		}
	}
	return true
}

func solveMapFieldType(fd *model.FieldDescriptor, localFD *model.FileDescriptor, globalFD *model.FileDescriptor) bool {
	rawstr := fd.RawFieldType
	if !strings.HasSuffix(rawstr, ">") {
		log.Errorf("map type format error, expected 'map<K,V>', got: '%s'", rawstr)
		return false
	}

	inner := rawstr[4 : len(rawstr)-1]
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		log.Errorf("map type format error, expected 'map<K,V>', got: '%s'", rawstr)
		return false
	}

	keyType := strings.TrimSpace(parts[0])
	valueType := strings.TrimSpace(parts[1])

	// 解析key类型
	if ft, ok := model.ParseFieldType(keyType); ok {
		fd.MapKeyType = ft
	} else {
		// 在本地和全局查找枚举类型
		desc := findDescriptor(localFD, globalFD, keyType)
		if desc == nil || desc.Kind != model.DescriptorKind_Enum {
			log.Errorf("map key type must be builtin type or enum, got: '%s'", keyType)
			return false
		}
		fd.MapKeyType = model.FieldType_Enum
		fd.MapKeyComplex = desc
	}

	// 解析value类型
	if ft, ok := model.ParseFieldType(valueType); ok {
		fd.MapValueType = ft
	} else {
		// 在本地和全局查找枚举或结构体类型
		desc := findDescriptor(localFD, globalFD, valueType)
		if desc == nil {
			log.Errorf("map value type not found: '%s'", valueType)
			return false
		}
		fd.MapValueComplex = desc
		switch desc.Kind {
		case model.DescriptorKind_Struct:
			fd.MapValueType = model.FieldType_Struct
		case model.DescriptorKind_Enum:
			fd.MapValueType = model.FieldType_Enum
		}
	}

	fd.RawFieldType = "" // 清空，避免重复解析
	return true
}

func findDescriptor(localFD *model.FileDescriptor, globalFD *model.FileDescriptor, name string) *model.Descriptor {
	if desc, ok := localFD.DescriptorByName[name]; ok {
		return desc
	}
	if desc, ok := globalFD.DescriptorByName[name]; ok {
		return desc
	}
	return nil
}

func filterFields(cachedFile map[string]*File, g *printer.Globals) {
	for _, v := range cachedFile {
		for _, vv := range v.LocalFD.Descriptors {
			for i := 0; i < len(vv.Fields); i++ {
				field := vv.Fields[i]
				field.Order = int32(i)
				if field.Meta == nil || g.FieldMark == "" {
					continue
				}
				mark := field.Meta.KVPair.GetString("Mark")
				if mark == "" || mark == g.FieldMark {
					continue
				}
				delete(vv.FieldByName, field.Name)
				delete(vv.FieldByNumber, field.EnumValue)
				vv.Fields = append(vv.Fields[:i], vv.Fields[i+1:]...)
				i--
			}
		}
	}
	for _, vv := range g.FileDescriptor.Descriptors {
		for i := 0; i < len(vv.Fields); i++ {
			field := vv.Fields[i]
			if field.Meta == nil || g.FieldMark == "" {
				continue
			}
			mark := field.Meta.KVPair.GetString("Mark")
			if mark == "" || mark == g.FieldMark {
				continue
			}
			vv.Fields = append(vv.Fields[:i], vv.Fields[i+1:]...)
			i--
		}
	}
}

func delNotPrintDef(cachedFile map[string]*File, g *printer.Globals) {
	delDefs := make(map[string]*model.Descriptor)
	for _, v := range cachedFile {
		for idx, vv := range v.LocalFD.Descriptors {
			if vv.NotPrint {
				v.LocalFD.Descriptors = append(v.LocalFD.Descriptors[:idx], v.LocalFD.Descriptors[idx+1:]...)
				delete(v.LocalFD.DescriptorByName, vv.Name)
				delDefs[vv.Name] = vv
			}
		}
	}
	for idx, vv := range g.FileDescriptor.Descriptors {
		if vv.NotPrint {
			g.FileDescriptor.Descriptors = append(g.FileDescriptor.Descriptors[:idx], g.FileDescriptor.Descriptors[idx+1:]...)
			delete(g.FileDescriptor.DescriptorByName, vv.Name)
			delDefs[vv.Name] = vv
		}
	}
	for _, v := range cachedFile {
		for _, vv := range v.LocalFD.Descriptors {
			for _, field := range vv.Fields {
				if field.Complex == nil {
					continue
				}
				_, ok := delDefs[field.Complex.Name]
				if !ok {
					continue
				}
				field.Type = model.FieldType_Int32
				field.Complex = nil
			}
		}
	}
	for _, vv := range g.FileDescriptor.Descriptors {
		for _, field := range vv.Fields {
			if field.Complex == nil {
				continue
			}
			_, ok := delDefs[field.Complex.Name]
			if !ok {
				continue
			}
			field.Type = model.FieldType_Int32
			field.Complex = nil
		}
	}
}

func Run(g *printer.Globals) bool {

	if !g.PreExport() {
		return false
	}

	cachedFile := cacheFile(g)

	fileObjList := make([]*File, 0)

	// 第一阶段：收集所有文件的类型定义
	for _, in := range g.InputFileList {

		inputFile := in.(string)

		mergeFileList := strings.Split(inputFile, "+")

		for index, fileName := range mergeFileList {

			file, _ := cachedFile[fileName]

			if file == nil {
				return false
			}

			file.GlobalFD = g.FileDescriptor

			// 只解析类型信息
			if !file.ExportTypes() {
				return false
			}

			// 主文件才写入全局信息
			if index == 0 {

				// 整合类型信息
				if !g.AddTypes(file.LocalFD) {
					return false
				}

				fileObjList = append(fileObjList, file)
			}
		}
	}

	// 第二阶段：解析所有文件的表头（此时所有类型都已收集完毕）
	for _, in := range g.InputFileList {

		inputFile := in.(string)

		mergeFileList := strings.Split(inputFile, "+")

		var mainMergeFile *File

		for index, fileName := range mergeFileList {

			file, _ := cachedFile[fileName]

			if file == nil {
				return false
			}

			// 解析表头
			if !file.ExportHeaders(mainMergeFile) {
				return false
			}

			if index == 0 {
				mainMergeFile = file
			} else {
				// 添加子文件
				mainMergeFile.mergeList = append(mainMergeFile.mergeList, file)
			}
		}
	}

	// 第二阶段后：将行类型结构体添加到全局 FileDescriptor
	for _, file := range fileObjList {
		if !g.AddTypes(file.LocalFD) {
			return false
		}
	}

	// 延迟解析map类型
	for _, file := range fileObjList {
		if !solveMapFieldTypes(file, g.FileDescriptor) {
			return false
		}
	}

	log.Infof("==========%s==========", i18n.String(i18n.Run_ExportSheetData))

	for _, file := range fileObjList {

		log.Infoln(filepath.Base(file.FileName))

		dataModel := model.NewDataModel(g.FieldMark)

		tab := model.NewTable()
		tab.LocalFD = file.LocalFD

		// 主表
		if !file.ExportData(dataModel, nil) {
			return false
		}

		// 子表提供数据
		for _, mergeFile := range file.mergeList {

			log.Infoln(filepath.Base(mergeFile.FileName), "--->", filepath.Base(file.FileName))

			// 电子表格数据导出到Table对象
			if !mergeFile.ExportData(dataModel, file.Header) {
				return false
			}
		}

		// 合并所有值到node节点
		if !mergeValues(dataModel, tab, file) {
			return false
		}

		// 整合类型信息和数据
		if !g.AddContent(tab) {
			return false
		}

	}
	filterFields(cachedFile, g)

	// 根据各种导出类型, 调用各导出器导出
	bPrint := g.PrintDataFile()
	if !bPrint {
		return false
	}
	delNotPrintDef(cachedFile, g)
	return g.PrintCodeFile()
}
