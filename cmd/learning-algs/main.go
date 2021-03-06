package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/britojr/scripts/cmd"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
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

	ds, err := filepath.Glob(inpDir + "/data/*.train")
	errchk.Check(err, "")
	cmd.RunCmd("mkdir "+outDir+" -p", 0)

	for _, d := range ds {
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		switch {
		case strings.Contains(algExec, cmd.AlgLibra):
			if len(algSub) == 0 {
				fmt.Printf("libra subcommand required\n")
				flag.PrintDefaults()
				os.Exit(1)
			}
			commandLibra(inpDir, outDir, algExec, algSub, name, timeOut)
		case strings.Contains(algExec, cmd.AlgLSDD):
			commandLSDD(inpDir, outDir, algExec, name, timeOut)
		case strings.Contains(algExec, cmd.AlgGobnilp):
			if len(parFile) == 0 {
				fmt.Printf("parameter file required\n")
				flag.PrintDefaults()
				os.Exit(1)
			}
			commandGobnilp(inpDir, outDir, algExec, parFile, name, timeOut)
		case strings.Contains(algExec, cmd.AlgBI):
			commandBI(inpDir, outDir, algExec, name, timeOut)
		default:
			fmt.Printf("alg %s not supported\n", algExec)
			os.Exit(1)
		}
	}
}

func commandLibra(inpDir, outDir, algExec, algSub, name string, timeOut int) {
	dataFile := inpDir + "/data/" + name
	outFile := outDir + "/" + name + "-" + algSub
	ext := "ac"
	arg := ""
	cmdstr := ""
	var err error
	cmd.RunCmd(fmt.Sprintf("rm -rf %s.out", outFile), 0)
	cmd.RunCmd(fmt.Sprintf("rm -rf %s.score", outFile), 0)
	switch algSub {
	case cmd.AlgSubCl, cmd.AlgSubBN:
		ext = "bn"
		arg = "-prior 1"
		fallthrough
	case cmd.AlgSubACBN, cmd.AlgSubACMN:
		cmd.RunCmd(fmt.Sprintf("rm -rf %s.%s", outFile, ext), 0)
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.%s %s -s %s.schema -log %s.out", algExec, algSub, dataFile, outFile, ext, arg, dataFile, outFile,
		)
		_, err = cmd.RunCmd(cmdstr, timeOut)
	case cmd.AlgSubSPN, cmd.AlgSubMT:
		cmd.RunCmd(fmt.Sprintf("rm -rf %s.spn", outFile), 0)
		cmd.RunCmd(fmt.Sprintf("rm -rf %s.ac", outFile), 0)
		seed := time.Now().UnixNano()
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.spn -seed %v %s.schema -log %s.out", algExec, algSub, dataFile, outFile, seed, dataFile, outFile,
		)
		_, err = cmd.RunCmd(cmdstr, timeOut)
		if err == nil {
			cmdstr = fmt.Sprintf(
				"%s spn2ac -m %s.spn -o %s.ac", algExec, outFile, outFile,
			)
			_, err = cmd.RunCmd(cmdstr, 0)
		}
	}
	if err == nil {
		cmdstr = fmt.Sprintf(
			"%s mscore -m %s.%s -i %s.test -log %s.score", algExec, outFile, ext, dataFile, outFile,
		)
		cmd.RunCmd(cmdstr, 0)
	}
}

func commandLSDD(inpDir, outDir, algExec, name string, timeOut int) {
	soluDir := outDir + "/" + name
	cmd.RunCmd("mkdir "+soluDir+" -p", 0)
	dataFile := inpDir + "/data/" + name
	cmdstr := fmt.Sprintf(
		"java -jar %s learn %s.train %s.valid %s", algExec, dataFile, dataFile, soluDir,
	)
	cmd.RunCmd(cmdstr, timeOut)
}

func commandGobnilp(inpDir, outDir, algExec, parFile, name string, timeOut int) {
	parms := createGobFile(parFile, outDir, name)
	cmdstr := fmt.Sprintf("%s -g=%s -f=dat %s/data/%s.train", algExec, parms, inpDir, name)
	cmd.RunCmd(cmdstr, timeOut)
}

func createGobFile(fname, outDir, dname string) string {
	r := ioutl.OpenFile(fname)
	w := ioutl.CreateFile(outDir + dname + ".set")
	_, err := io.Copy(w, r)
	errchk.Check(err, "")
	fmt.Fprintf(w, "\ngobnilp/outputfile/solution = \"%s/<probname>.solution\"", outDir)
	fmt.Fprintf(w, "\ngobnilp/outputfile/scoreandtime = \"%s/<probname>.times\"\n", outDir)
	return w.Name()
}

func commandBI(inpDir, outDir, algExec, name string, timeOut int) {
	soluDir := outDir + "/" + name
	cmd.RunCmd("mkdir "+soluDir+" -p", 0)
	dataFile := inpDir + "/data/" + name
	// unlike the others, BI files must have a header and specific extension
	trainFile, testFile := dataFile+"-train.arff", dataFile+"-test.arff"
	hdr := createHeader(dataFile)
	copyWithHeader(trainFile, dataFile+".train", hdr)
	copyWithHeader(testFile, dataFile+".test", hdr)
	// defer os.Remove(trainFile)
	// defer os.Remove(testFile)

	cmdstr := fmt.Sprintf(
		"java -Xmx2G -cp %s clustering/LearnAndTest %s %s %s", algExec, trainFile, testFile, soluDir,
	)
	cmd.RunCmd(cmdstr, timeOut)
}

func commandEAST(inpDir, outDir, algExec, name string, timeOut int) {
	// 	Arguments:
	// 	Setting for screening stage:
	// 		args[0]: Number of starting points of local EM
	// 		args[1]: Number of continued steps of local EM
	// 		args[2]: Convergence threshold in loglikelihood
	// 	Setting for evaluation stage:
	// 		args[3]: Maximum number of candidate models to enter evaluation stage
	// 		args[4]: Number of starting points of local EM
	// 		args[5]: Number of continued steps of local EM
	// 		args[6]: Convergence threshold in loglikelihood
	// 	Setting for parameter optimization:
	// 		args[7]: Number of starting points of full EM
	// 		args[8]: Number of maximum steps of full EM
	// 		args[9]: Convergence threshold in loglikelihood
	// 	General setting:
	// 		args[10]: Path to data file (see 5k.data for format)
	// 		args[11]: Path to ouput directory
	// 		args[12]: Path to initial model (optional)
	// 		args[13]: Conduct adjustment for initial model first or not (true/false, optional)
	//
	// Example: $ java -Xmx1024M -cp east.jar EAST 4 10 0.1 50 16 20 0.1 32 100 0.1 5k.data . >& ./log.txt
	args := "4 10 0.1 50 16 20 0.1 32 100 0.1"

	// TODO: needs to convert to EAST format
	soluDir := outDir + "/" + name
	cmd.RunCmd("mkdir "+soluDir+" -p", 0)
	dataFile := inpDir + "/data/" + name

	cmdstr := fmt.Sprintf(
		"java -Xmx2G -cp %s EAST %s %s %s", algExec, args, dataFile, soluDir,
	)
	cmd.RunCmd(cmdstr, timeOut)
}

func copyWithHeader(dst, src, hdr string) {
	r := ioutl.OpenFile(src)
	w := ioutl.CreateFile(dst)
	fmt.Fprintln(w, hdr)
	_, err := io.Copy(w, r)
	errchk.Check(err, "")
}

func createHeader(baseName string) string {
	line := ""
	fh := ioutl.OpenFile(baseName + ".hdr")
	fmt.Fscanln(fh, &line)
	fh.Close()
	names := strings.Split(line, ",")
	fs := ioutl.OpenFile(baseName + ".schema")
	fmt.Fscanln(fs, &line)
	fs.Close()
	cards := strings.Split(line, ",")
	hdr := "@relation data\n"
	for i, name := range names {
		states := make([]string, conv.Atoi(cards[i]))
		for j := range states {
			states[j] = strconv.Itoa(j)
		}
		hdr += fmt.Sprintf("@attribute %s {%s}\n", name, strings.Join(states, ","))
	}
	hdr += "@data\n"
	return hdr
}
