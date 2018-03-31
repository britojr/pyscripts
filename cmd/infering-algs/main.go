package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/britojr/scripts/cmd"
	"github.com/britojr/utl/errchk"
)

func main() {
	var (
		inpDir, outDir  string
		algExec, algSub string
		parFile         string
		timeOut         int
	)
	flag.StringVar(&inpDir, "i", "", "input directory")
	flag.StringVar(&outDir, "o", "", "output directory")
	flag.StringVar(&algExec, "e", "", "executable file")
	flag.StringVar(&algSub, "s", "", "subcommand libra:(cl|bnlearn|acbn|acmn|idspn|mtlearn)")
	flag.StringVar(&parFile, "p", "", "parameter file")
	flag.IntVar(&timeOut, "t", 0, "timeout in seconds")
	flag.Parse()

	if len(inpDir) == 0 || len(outDir) == 0 || len(algExec) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	// remove prints
	fmt.Println("DIRS:", inpDir, outDir)
	fmt.Println("ALGS:", algExec, algSub)
	fmt.Println("TIME:", timeOut)

	ds, err := filepath.Glob(inpDir + "/query/*.q")
	errchk.Check(err, "")

	for _, d := range ds {
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		switch {
		case strings.Contains(algExec, cmd.AlgLibra):
			if len(algSub) == 0 {
				fmt.Printf("libra subcommand required\n")
				flag.PrintDefaults()
				os.Exit(1)
			}
			inferLibra(inpDir, outDir, algExec, algSub, name, timeOut)
		case strings.Contains(algExec, cmd.AlgLSDD):
		case strings.Contains(algExec, cmd.AlgGobnilp):
		case strings.Contains(algExec, cmd.AlgBI):
		default:
			fmt.Printf("alg %s not supported\n", algExec)
			os.Exit(1)
		}
	}
}

func inferLibra(inpDir, outDir, algExec, algSub, name string, timeOut int) {
	dataFile := inpDir + "/data/" + name
	outFile := outDir + "/" + name + "-" + algSub
	ext := "ac"
	arg := ""
	cmdstr := ""
	switch algSub {
	case cmd.AlgSubCl, cmd.AlgSubBN:
		ext = "bn"
		arg = "-prior 1"
		fallthrough
	case cmd.AlgSubACBN, cmd.AlgSubACMN:
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.%s %s -log %s.out", algExec, algSub, dataFile, outFile, ext, arg, outFile,
		)
		cmd.RunCmd(cmdstr, timeOut)
	case cmd.AlgSubSPN, cmd.AlgSubMT:
		seed := time.Now().UnixNano()
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.spn -seed %v -log %s.out", algExec, algSub, dataFile, outFile, seed, outFile,
		)
		cmd.RunCmd(cmdstr, timeOut)
		cmdstr = fmt.Sprintf(
			"%s spn2ac -m %s.spn -o %s.ac", algExec, outFile, outFile,
		)
		fmt.Println(cmdstr)
		_, err := cmd.ExecCmd(cmdstr)
		errchk.Check(err, "")
	}
	cmdstr = fmt.Sprintf(
		"%s mscore -m %s.%s -i %s.test -log %s.score", algExec, outFile, ext, dataFile, outFile,
	)
	fmt.Println(cmdstr)
	_, err := cmd.ExecCmd(cmdstr)
	errchk.Check(err, "")
}
