package main

import (
	"flag"
	"fmt"
)

func main() {
	var inpDir, outDir string
	flag.StringVar(&inpDir, "i", "", "input directory")
	flag.StringVar(&outDir, "o", "", "output directory")
	flag.Parse()

	if len(inpDir) == 0 || len(outDir) == 0 {
		flag.PrintDefaults()
	}

	fmt.Println(inpDir, outDir)
}
