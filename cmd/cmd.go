package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
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
