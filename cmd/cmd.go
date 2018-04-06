package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/britojr/bnutils/bif"
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

func RunCmd(cmdstr string, timeOut int) ([]byte, error) {
	fmt.Println(cmdstr)
	out, err := ExecCmdTimeout(cmdstr, timeOut)
	if err != nil {
		if err == ErrTimeout {
			fmt.Println(err)
		} else {
			fmt.Printf("command errored: %v\n", err)
			fmt.Println(string(out))
		}
	}
	return out, err
}

func ExecCmdTimeout(cmdstr string, t int) ([]byte, error) {
	if t <= 0 {
		return ExecCmd(cmdstr)
	}
	return ExecCmd(fmt.Sprintf("timeout %v %s", t, cmdstr))
}
func ExecCmdTimeout1(cmdstr string, t int) ([]byte, error) {
	if t <= 0 {
		return ExecCmd(cmdstr)
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
func ExecCmd(cmdstr string) ([]byte, error) {
	args := strings.Fields(cmdstr)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return out, err
}

func DataAppend(netFile, inpFile, outFile string) {
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
