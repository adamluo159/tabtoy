package printer

type PrinterContext struct {
	outFile string
	p       Printer
	name    string
}

func (self *PrinterContext) Start(g *Globals) bool {

	if self.outFile != "" {
		log.Infof("[%s] %s", self.name, self.outFile)
	}

	bf := self.p.Run(g)

	if bf == nil {
		return false
	}

	// 当使用-json_dir参数且当前是JSON打印机时，跳过写入合并的JSON文件
	if self.name == "json" && g.JsonDir != "" {
		return true
	}

	return bf.WriteFile(self.outFile) == nil
}

type Printer interface {
	Run(g *Globals) *Stream
}

var printerByExt = make(map[string]Printer)

func RegisterPrinter(ext string, p Printer) {

	if _, ok := printerByExt[ext]; ok {
		panic("duplicate printer")
	}

	printerByExt[ext] = p
}
