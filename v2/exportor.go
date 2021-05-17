package v2

import (
	"path/filepath"
	"strings"

	"github.com/adamluo159/tabtoy/v2/i18n"
	"github.com/adamluo159/tabtoy/v2/model"
	"github.com/adamluo159/tabtoy/v2/printer"
)

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

	log.Infof("==========%s==========", i18n.String(i18n.Run_CollectTypeInfo))

	// 合并类型
	for _, in := range g.InputFileList {

		inputFile := in.(string)

		var mainMergeFile *File

		mergeFileList := strings.Split(inputFile, "+")

		for index, fileName := range mergeFileList {

			file, _ := cachedFile[fileName]

			if file == nil {
				return false
			}

			var mergeTarget string
			if len(mergeFileList) > 1 {
				mergeTarget = "--> " + filepath.Base(mergeFileList[0])
			}

			log.Infoln(filepath.Base(fileName), mergeTarget)

			file.GlobalFD = g.FileDescriptor

			// 电子表格数据导出到Table对象
			if !file.ExportLocalType(mainMergeFile) {
				return false
			}

			// 主文件才写入全局信息
			if index == 0 {

				// 整合类型信息和数据
				if !g.AddTypes(file.LocalFD) {
					return false
				}

				// 只写入主文件的文件列表
				if file.Header != nil {

					fileObjList = append(fileObjList, file)
				}

				mainMergeFile = file
			} else {

				// 添加自文件
				mainMergeFile.mergeList = append(mainMergeFile.mergeList, file)

			}

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
