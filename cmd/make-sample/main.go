package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/britojr/scripts/cmd"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	var (
		inpDir                string
		nTrain, nTest, nValid int
	)
	flag.StringVar(&inpDir, "i", "", "input directory")
	flag.IntVar(&nTrain, "tr", 0, "number of samples for training file")
	flag.IntVar(&nTest, "te", 0, "number of samples for testing file")
	flag.IntVar(&nValid, "va", 0, "number of samples for validation file")
	flag.Parse()

	if len(inpDir) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	exts := []string{"train", "test", "valid"}
	n := []int{nTrain, nTest, nValid}
	cmd.RunCmd(fmt.Sprintf("mkdir %s/data/ -p", inpDir), 0)
	cmd.RunCmd(fmt.Sprintf("mkdir %s/model/ -p", inpDir), 0)
	ds, err := filepath.Glob(inpDir + "/*.bif")
	errchk.Check(err, "")
	for _, d := range ds {
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		createSchema(inpDir, name)
		for i, ext := range exts {
			cmdstr := fmt.Sprintf(
				"libra bnsample -m %s -o %s/data/%s.%s -n %d -seed %d",
				d, inpDir, name, ext, n[i], randSource.Int31(),
			)
			out, err := cmd.RunCmd(cmdstr, 0)
			errchk.Check(err, string(out))
		}
		cmdstr := fmt.Sprintf(
			"libra mscore -m %s -i %s/data/%s.test -log %s/model/%s.score",
			d, inpDir, name, inpDir, name,
		)
		out, err := cmd.RunCmd(cmdstr, 0)
		errchk.Check(err, string(out))
	}
}

func createSchema(inpDir, name string) {
	cmdstr := fmt.Sprintf("libra fstats -i %s/%s.bif", inpDir, name)
	out, err := cmd.RunCmd(cmdstr, 0)
	errchk.Check(err, string(out))
	schema, hdr := "Schema: ", ""
	for _, line := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
		if len(line) > len(schema) && line[:len(schema)] == schema {
			hdr = line[len(schema):]
		}
	}
	f := ioutl.CreateFile(inpDir + "/data/" + name + ".schema")
	fmt.Fprintf(f, "%s\n", hdr)
	f.Close()
}
