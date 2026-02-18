package main

import (
	"os"

	"github.com/rickhallett/darkfactorio/internal/dfgatecli"
)

func main() {
	os.Exit(dfgatecli.Run(os.Args[1:]))
}

