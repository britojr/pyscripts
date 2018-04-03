package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/scripts/cmd"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	"github.com/kniren/gota/dataframe"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	var (
		inpDir, outDir string
		nInternals     int
	)
	flag.StringVar(&inpDir, "i", "", "input directory")
	flag.StringVar(&outDir, "o", "", "output directory")
	flag.IntVar(&nInternals, "n", 0, "number of variables to hide")
	flag.Parse()

	if len(inpDir) == 0 || len(outDir) == 0 || nInternals == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	ds, err := filepath.Glob(inpDir + "/*.bif")
	errchk.Check(err, "")
	cmd.RunCmd(fmt.Sprintf("mkdir %s/data/ -p", outDir), 0)
	cmd.RunCmd(fmt.Sprintf("mkdir %s/query/ -p", outDir), 0)
	for _, d := range ds {
		b := bif.ParseStruct(d)
		xs := sampleInternals(b, nInternals)
		name := strings.TrimSuffix(filepath.Base(d), filepath.Ext(filepath.Base(d)))
		createCutFiles(inpDir, outDir, name, xs)
	}
}

func sampleInternals(b *bif.Struct, n int) (xs []int) {
	is := b.Internals()
	perm := randSource.Perm(len(is))
	if n > len(is) {
		n = len(is)
	}
	for _, v := range perm[:n] {
		xs = append(xs, is[v].ID())
	}
	return
}

func createCutFiles(inpDir, outDir, name string, cols []int) {
	i := 1
	for _, ext := range []string{"train", "test", "valid", "schema"} {
		cutFile(
			fmt.Sprintf("%s/data/%s.%s", inpDir, name, ext),
			fmt.Sprintf("%s/data/%s-X%dI%d.%s", outDir, name, len(cols), i, ext),
			cols,
		)
	}
	for _, ext := range []string{"q", "ev"} {
		cutFile(
			fmt.Sprintf("%s/query/%s.%s", inpDir, name, ext),
			fmt.Sprintf("%s/query/%s-X%dI%d.%s", outDir, name, len(cols), i, ext),
			cols,
		)
	}
}

func cutFile(fi, fo string, cols []int) {
	fmt.Println(fo)
	df := dataframe.ReadCSV(ioutl.OpenFile(fi), dataframe.HasHeader(false)).Drop(cols)
	err := df.WriteCSV(ioutl.CreateFile(fo), dataframe.WriteHeader(false))
	errchk.Check(err, "")
}
