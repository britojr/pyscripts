package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

// suported algorithms
const (
	AlgLibra   = "libra"
	AlgGobnilp = "gobnilp"
	AlgLSDD    = "LearnSDD"
	AlgBI      = "BI"
)

// suported subcommands
const (
	AlgSubBN   = "bnlearn"
	AlgSubCl   = "cl"
	AlgSubACBN = "acbn"
	AlgSubACMN = "acmn"
	AlgSubSPN  = "idspn"
	AlgSubMT   = "mtlearn"
)

var ErrTimeout error = errTimeOut{}

type errTimeOut struct{}

func (errTimeOut) Error() string { return "command timed out" }

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
	_, err = execCmd("mkdir " + outDir + " -p")
	errchk.Check(err, "")

	for _, d := range ds {
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		switch {
		case strings.Contains(algExec, AlgLibra):
			if len(algSub) == 0 {
				fmt.Printf("libra subcommand required\n")
				flag.PrintDefaults()
				os.Exit(1)
			}
			commandLibra(inpDir, outDir, algExec, algSub, name, timeOut)
		case strings.Contains(algExec, AlgLSDD):
			commandLSDD(inpDir, outDir, algExec, name, timeOut)
		case strings.Contains(algExec, AlgGobnilp):
			if len(parFile) == 0 {
				fmt.Printf("parameter file required\n")
				flag.PrintDefaults()
				os.Exit(1)
			}
			commandGobnilp(inpDir, outDir, algExec, parFile, name, timeOut)
		case strings.Contains(algExec, AlgBI):
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
	switch algSub {
	case AlgSubCl, AlgSubBN:
		ext = "bn"
		arg = "-prior 1"
		fallthrough
	case AlgSubACBN, AlgSubACMN:
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.%s %s -log %s.out", algExec, algSub, dataFile, outFile, ext, arg, outFile,
		)
		runCmd(cmdstr, timeOut)
	case AlgSubSPN, AlgSubMT:
		seed := time.Now().UnixNano()
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.spn -seed %v -log %s.out", algExec, algSub, dataFile, outFile, seed, outFile,
		)
		runCmd(cmdstr, timeOut)
		cmdstr = fmt.Sprintf(
			"%s spn2ac -m %s.spn -o %s.ac", algExec, outFile, outFile,
		)
		fmt.Println(cmdstr)
		_, err := execCmd(cmdstr)
		errchk.Check(err, "")
	}
	cmdstr = fmt.Sprintf(
		"%s mscore -m %s.%s -i %s.test -log %s.score", algExec, outFile, ext, dataFile, outFile,
	)
	fmt.Println(cmdstr)
	_, err := execCmd(cmdstr)
	errchk.Check(err, "")
}

func commandLSDD(inpDir, outDir, algExec, name string, timeOut int) {
	soluDir := outDir + "/" + name
	_, err := execCmd("mkdir " + soluDir + " -p")
	errchk.Check(err, "")
	dataFile := inpDir + "/data/" + name
	cmdstr := fmt.Sprintf(
		"java -jar %s learn %s.train %s.valid %s", algExec, dataFile, dataFile, soluDir,
	)
	runCmd(cmdstr, timeOut)
}

func commandGobnilp(inpDir, outDir, algExec, parFile, name string, timeOut int) {
	parms := createGobFile(parFile, outDir, name)
	cmdstr := fmt.Sprintf("%s -g=%s -f=dat %s/data/%s.train", algExec, parms, inpDir, name)
	runCmd(cmdstr, timeOut)
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
	_, err := execCmd("mkdir " + soluDir + " -p")
	errchk.Check(err, "")
	dataFile := inpDir + "/data/" + name
	// unlike the others, BI files must have a header and specific extension
	trainFile, testFile := dataFile+"-train.csv", dataFile+"-test.csv"
	copyWithHeader(trainFile, dataFile+".train")
	copyWithHeader(testFile, dataFile+".test")
	defer os.Remove(trainFile)
	defer os.Remove(testFile)

	cmdstr := fmt.Sprintf(
		"java -Xmx2G -cp %s clustering/LearnAndTest %s %s %s", algExec, trainFile, testFile, soluDir,
	)
	runCmd(cmdstr, timeOut)
}

func copyWithHeader(dst, src string) {
	line := ""
	r := ioutl.OpenFile(src)
	fmt.Fscanln(r, &line)
	hdr := make([]string, len(strings.Split(line, ",")))
	for i := range hdr {
		hdr[i] = "x" + strconv.Itoa(i)
	}
	w := ioutl.CreateFile(dst)
	fmt.Fprintln(w, strings.Join(hdr, ","))
	fmt.Fprintln(w, line)
	_, err := io.Copy(w, r)
	errchk.Check(err, "")
}

func runCmd(cmdstr string, timeOut int) {
	fmt.Println(cmdstr)
	out, err := execCmdTimeout(cmdstr, timeOut)
	if err != nil {
		if err == ErrTimeout {
			fmt.Println(err)
		} else {
			fmt.Printf("command errored: %v\n", err)
			fmt.Println(string(out))
		}
	}
}

func execCmdTimeout(cmdstr string, t int) ([]byte, error) {
	if t <= 0 {
		return execCmd(cmdstr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t)*time.Second)
	defer cancel()

	args := strings.Fields(cmdstr)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return out, ErrTimeout
	}
	return out, err
}
func execCmd(cmdstr string) ([]byte, error) {
	args := strings.Fields(cmdstr)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return out, err
}
