package build

import (
	"fmt"
	"time"
)

var (
	Version   string = time.Now().Format("20060102150405")
	GitCommit string
	BuildTime string
)

func Print() {
	fmt.Println("Version: ", Version)
	fmt.Println("GitCommit: ", GitCommit)
	fmt.Println("BuildTime: ", BuildTime)
}
