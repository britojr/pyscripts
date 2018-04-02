package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/britojr/scripts/cmd"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	"github.com/britojr/utl/stats"
)

func main() {
	var (
		inpDir, outDir string
		algExec        string
	)
	flag.StringVar(&inpDir, "i", "", "input directory")
	flag.StringVar(&outDir, "o", "", "output directory")
	flag.StringVar(&algExec, "e", "", "executable file")
	flag.Parse()

	if len(inpDir) == 0 || len(outDir) == 0 || len(algExec) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	// remove prints
	fmt.Println("DIRS:", inpDir, outDir)
	fmt.Println("ALGS:", algExec)

	ds, err := filepath.Glob(inpDir + "/query/*.infkey")
	errchk.Check(err, "")

	for _, d := range ds {
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		switch {
		case strings.Contains(algExec, cmd.AlgLibra):
			calcASELibra(inpDir, outDir, name, "exact")
		case strings.Contains(algExec, cmd.AlgLSDD):
		case strings.Contains(algExec, cmd.AlgGobnilp):
		case strings.Contains(algExec, cmd.AlgBI):
		default:
			fmt.Printf("alg %s not supported\n", algExec)
			os.Exit(1)
		}
	}
}

func calcASELibra(inpDir, outDir, name, ext string) {
	keyFile := inpDir + "/query/" + name + ".infkey"
	infFiles, err := filepath.Glob(outDir + "/" + name + "*." + ext)
	errchk.Check(err, "")
	keys := readInfFile(keyFile)
	for _, infFile := range infFiles {
		inf := readInfFile(infFile)
		if len(keys) == len(inf) {
			f := ioutl.CreateFile(infFile + ".ase")
			fmt.Fprintf(f, "%v\n", stats.MSE(keys, inf))
			f.Close()
		}
	}
}

func readInfFile(fname string) (vs []float64) {
	f := ioutl.OpenFile(fname)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		v, err := strconv.ParseFloat(scanner.Text(), 64)
		if err == nil {
			vs = append(vs, math.Exp(v))
		}
	}
	return
}
