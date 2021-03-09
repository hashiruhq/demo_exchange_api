package main // import "around25.com/exchange/demo_api"

import (
	"runtime"

	"around25.com/exchange/demo_api/cmd"
)

func init() {
	// set proc count
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	cmd.Execute()
}
