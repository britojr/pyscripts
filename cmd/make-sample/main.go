package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/britojr/bnutils/bif"
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
	b := bif.ParseStruct(inpDir + "/" + name + ".bif")
	vs := b.Variables()
	cards, names, maxs := make([]string, len(vs)), make([]string, len(vs)), make([]string, len(vs))
	for i, v := range vs {
		cards[i] = strconv.Itoa(v.NState())
		names[i] = v.Name()
		maxs[i] = strconv.Itoa(v.NState() - 1)
	}
	f := ioutl.CreateFile(inpDir + "/data/" + name + ".schema")
	f.Close()
	fh := ioutl.CreateFile(inpDir + "/data/" + name + ".hdr")
	fmt.Fprintf(fh, "%s\n", strings.Join(names, ","))
	fmt.Fprintf(fh, "%s\n", strings.Join(maxs, ","))
	fh.Close()
}
