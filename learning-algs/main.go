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
	AlgLSDD    = "LearnSDD"
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
	fmt.Println("ALGS:", algSub, algExec)
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
			"%s %s -i %s.train -o %s.%s %s &> %s.out", algExec, algSub, dataFile, outFile, ext, arg, outFile,
		)
		fmt.Println(cmdstr)
		_, err := execCmdTimeout(cmdstr, timeOut)
		if err != nil && err != ErrTimeout {
			fmt.Printf("command errored: %v\n", err)
		}
	case AlgSubSPN, AlgSubMT:
		// libra ${ALG} -i ${ROOT}/data/${NAME}.train -o ${OUT}/${NAME}-${ALG}.spn -seed ${SEED} &> ${OUT}/${NAME}-${ALG}.out
		seed := 0
		cmdstr = fmt.Sprintf(
			"%s %s -i %s.train -o %s.spn -seed %v &> %s.out", algExec, algSub, dataFile, outFile, seed, outFile,
		)
		fmt.Println(cmdstr)
		_, err := execCmdTimeout(cmdstr, timeOut)
		if err != nil && err != ErrTimeout {
			fmt.Printf("command errored: %v\n", err)
		}
		// libra spn2ac -m ${OUT}/${NAME}-${ALG}.spn -o ${OUT}/${NAME}-${ALG}.ac &>> ${OUT}/${NAME}-${ALG}.out
		cmdstr = fmt.Sprintf(
			"%s spn2ac -m %s.spn -o %s.ac &>> %s.out", algExec, outFile, outFile, outFile,
		)
		fmt.Println(cmdstr)
		_, err = execCmd(cmdstr)
		errchk.Check(err, "")
	}
	// libra mscore -m ${OUT}/${NAME}-${ALG}.${EXT} -i ${ROOT}/data/${NAME}.test &> ${OUT}/${NAME}-${ALG}.score
	cmdstr = fmt.Sprintf(
		"%s mscore -m %s.%s -i %s.test &> %s.score", algExec, outFile, ext, dataFile, outFile,
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

	fmt.Println(cmdstr)
	_, err = execCmdTimeout(cmdstr, timeOut)
	if err != nil && err != ErrTimeout {
		fmt.Printf("command errored: %v\n", err)
	}
}

func commandGobnilp(inpDir, outDir, algExec, parFile, name string, timeOut int) {
	parms := createGobFile(parFile, outDir, name)
	cmdstr := fmt.Sprintf("%s -g=%s -f=dat %s/data/%s.train", algExec, parms, inpDir, name)

	fmt.Println(cmdstr)
	_, err := execCmdTimeout(cmdstr, timeOut)
	if err != nil && err != ErrTimeout {
		fmt.Printf("command errored: %v\n", err)
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
