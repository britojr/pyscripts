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
		fmt.Println(name)
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
			inferBI(inpDir, outDir, name, timeOut)
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

func inferBI(inpDir, outDir, name string, timeOut int) {
	// convert actual bif format respecting the original variable ordering
	// ctconv -i exp/medium/bi/alarm/FinalBIModel.bif -o exp/medium/bi/alarm.bif -t bi-bif -d exp/medium/data/alarm.hdr
	baseFile := outDir + "/" + name
	dataFile := inpDir + "/data/" + name
	queryFile := inpDir + "/query/" + name
	cmd.RunCmd(fmt.Sprintf(
		"ctconv -t bi-bif -i %s/FinalBIModel.bif -o %s.bif -d %s.hdr",
		baseFile, baseFile, dataFile,
	), 0)
	// add missing values corresponding to the hidden variables
	cmd.DataAppend(baseFile+".bif", queryFile+".q", baseFile+".q")
	cmd.DataAppend(baseFile+".bif", queryFile+".ev", baseFile+".ev")
	// convert to libra ac and perform the query
	cmd.RunCmd(fmt.Sprintf("libra acve -m %s.bif -o %s.ac", baseFile, baseFile), 0)
	cmd.RunCmd(fmt.Sprintf(
		"libra acquery -m %s.ac -q %s.q -ev %s.ev -log %s.exact",
		baseFile, baseFile, baseFile, baseFile,
	), timeOut)
}
