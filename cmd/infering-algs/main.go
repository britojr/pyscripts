package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	cmdstr := ""
	ext := "ac"
	infcmd := "acquery"
	netFile := outDir + "/" + name + "-" + algSub
	queryFile := inpDir + "/query/" + name
	switch algSub {
	case cmd.AlgSubCl, cmd.AlgSubBN:
		// libra acve -m ${OUT}/${NAME}-${ALG}.bn -o ${OUT}/${NAME}-${ALG}.ac
		// libra fstats -i ${OUT}/${NAME}-${ALG}.ac &> ${OUT}/${NAME}-${ALG}.outac
		cmdstr = fmt.Sprintf(
			"%s acve -m %s.bn -o %s.ac", algExec, netFile, netFile,
		)
		cmd.RunCmd(cmdstr, timeOut)
	case cmd.AlgSubSPN, cmd.AlgSubMT:
		ext = "spn"
		infcmd = "spquery"
	}
	// libra ${CMD} -m ${OUT}/${NAME}-${ALG}.${EXT} -q ${ROOT}/query/${NAME}.q -ev ${ROOT}/query/${NAME}.ev &> ${OUT}/${NAME}-${ALG}.exact
	cmdstr = fmt.Sprintf(
		"%s %s -m %s.%s -q %s.q -ev %s.ev -log %s.exact", algExec, infcmd, netFile, ext, queryFile, queryFile, netFile,
	)
	cmd.RunCmd(cmdstr, timeOut)
}
