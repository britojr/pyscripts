package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/utl/ioutl"
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

	b := bif.ParseStruct(netFile)
	nLatent := len(b.Variables()) - getNumCol(inpFile)
	cols := make([]string, nLatent)
	for i := range cols {
		cols[i] = "*"
	}
	fmt.Println(len(cols))
	appendColToFile(inpFile, outFile, cols)
}

func getNumCol(fname string) int {
	r := ioutl.OpenFile(fname)
	defer r.Close()
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return len(strings.Split(scanner.Text(), ","))
}

func appendColToFile(src, dst string, cols []string) {
	fmt.Println(dst)
	r := ioutl.OpenFile(src)
	w := ioutl.CreateFile(dst)
	defer r.Close()
	defer w.Close()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Fprintf(w, "%s,%s\n", scanner.Text(), strings.Join(cols, ","))
	}
}
