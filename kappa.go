package main

import (
	"runtime"

	"github.com/subsilent/kappa/commands"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	commands.Execute()
}
