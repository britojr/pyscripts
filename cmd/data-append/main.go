package main

import (
	"flag"
	"os"

	"github.com/britojr/scripts/cmd"
)

func main() {
	var (
		netFile, inpFile, outFile string
	)
	flag.StringVar(&netFile, "n", "", "network file")
	flag.StringVar(&inpFile, "i", "", "input data file")
	flag.StringVar(&outFile, "o", "", "output data file")
	flag.Parse()

	if len(netFile) == 0 || len(inpFile) == 0 || len(outFile) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	cmd.DataAppend(netFile, inpFile, outFile)
}
