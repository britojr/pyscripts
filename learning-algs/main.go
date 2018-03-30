package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

// suported algorithms
const (
	AlgLibra   = "libra"
	AlgGobnilp = "gobnilp"
	AlgLSDD    = "lsdd"
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
		inpDir, outDir           string
		algName, algSub, algExec string
		parFile                  string
		timeOut                  int
	)
	flag.StringVar(&inpDir, "i", "", "input directory")
	flag.StringVar(&outDir, "o", "", "output directory")
	flag.StringVar(&algName, "a", "", "algorithm name (libra|lsdd|gobnilp)")
	flag.StringVar(&algSub, "s", "", "subcommand libra:(cl|bnlearn|acbn|acmn|idspn|mtlearn)")
	flag.StringVar(&algExec, "e", "", "executable file")
	flag.StringVar(&parFile, "p", "", "parameter file")
	flag.IntVar(&timeOut, "t", 0, "timeout in seconds")
	flag.Parse()

	if len(inpDir) == 0 || len(outDir) == 0 || len(algName) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	// remove prints
	fmt.Println("DIRS:", inpDir, outDir)
	fmt.Println("ALGS:", algName, algSub, algExec)
	fmt.Println("TIME:", timeOut)

	// common
	ds, err := filepath.Glob(inpDir + "/data/*.train")
	errchk.Check(err, "")
	_, err = execCmd("mkdir " + outDir + " -p")
	errchk.Check(err, "")

	// gobnilp
	if len(parFile) == 0 {
		fmt.Printf("parameter file required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	for _, d := range ds {
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		parms := createGobFile(parFile, outDir, name)
		cmdstr := fmt.Sprintf("%s -g=%s -f=dat %s/data/%s.train", algExec, parms, inpDir, name)
		fmt.Println(cmdstr)
		_, err := execCmdTimeout(cmdstr, timeOut)
		if err != nil && err != ErrTimeout {
			fmt.Printf("command errored: %v\n", err)
		}
	}
}

func createGobFile(fname, outDir, dname string) string {
	data, err := ioutil.ReadFile(fname)
	errchk.Check(err, "")
	content := string(data) +
		fmt.Sprintf("\n\ngobnilp/outputfile/solution = \"%s/<probname>.solution\"", outDir) +
		fmt.Sprintf("\ngobnilp/outputfile/scoreandtime = \"%s/<probname>.times\"\n", outDir)
	f := ioutl.CreateFile(outDir + dname + ".set")
	fmt.Fprintf(f, content)
	return f.Name()
}

func execCmdTimeout(cmdstr string, t int) ([]byte, error) {
	if t <= 0 {
		return execCmd(cmdstr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t)*time.Second)
	defer cancel()

	args := strings.Fields(cmdstr)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		return out, ErrTimeout
	}
	return out, err
}
func execCmd(cmdstr string) ([]byte, error) {
	args := strings.Fields(cmdstr)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	return out, err
}
