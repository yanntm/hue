// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dalzilio/hue/pkg/formula"
	"github.com/dalzilio/hue/pkg/hlnet"
	"github.com/dalzilio/hue/pkg/pnml"
	flag "github.com/spf13/pflag"
)

// builddate stores the compilation date for the executable, in %Y%m%d format.
// Set in the Makefile with the result of:
//
//	date -u +%d/%m/%Y
var builddate string = "2020/01/01"

// gitversion stores the current git version. Set in the makefile with the result
// of:
//
//	git describe --tags --dirty --always
var gitversion string = "v0"

func main() {
	var flaghelp = flag.BoolP("help", "h", false, "print this message")
	var flagstat = flag.BoolP("stat", "s", false, "print statistics information")
	var flagversion = flag.Bool("version", false, "print version number and generation date of hwalk")
	var flagshowqueries = flag.Bool("show-queries", false, "print queries on standard output")
	var flaghidetrivial = flag.Bool("hide-trivial", false, "hide results for trivial queries")
	var flaghideundef = flag.Bool("hide-undef", false, "hide queries with UNDEF result")
	var flagreach = flag.BoolP("reachability", "r", false, "check ReachabilityCardinality.xml file")
	var flagfire = flag.BoolP("fireability", "f", false, "check ReachabilityFireability.xml file")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "hwalk version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nfiles:\n")
		fmt.Fprintf(os.Stderr, "   outfile: output is always on stdout\n")
		fmt.Fprintf(os.Stderr, "   errorfile: error are always printed on stderr\n")
	}

	flag.Parse()
	needformulas := *flagfire || *flagreach
	N := len(flag.Args())

	if *flaghelp {
		flag.Usage()
		os.Exit(0)
	}

	if *flagversion {
		fmt.Printf("hue version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
	}

	switch {
	case N != 1:
		fmt.Println("should have exactly one MCC folder parameter")
		flag.Usage()
		os.Exit(1)
	}

	// we capture panics
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Error in generation: cannot compute")
			os.Exit(1)
		}
	}()

	start := time.Now()

	// we check the XML directory exists
	dirname := flag.Arg(0)
	xmlDirInfo, err := os.Stat(dirname)
	if err != nil {
		log.Println("Error with directory:", err)
		os.Exit(1)
		return
	}
	if !xmlDirInfo.IsDir() {
		log.Printf("Path %s is not a directory\n", dirname)
		os.Exit(1)
		return
	}

	// we try to open the model.xml file from it
	xmlFile, err := os.Open(filepath.Join(dirname, "model.pnml"))
	if err != nil {
		log.Println("Error opening file:", err)
		os.Exit(1)
		return
	}
	defer xmlFile.Close()

	decoder := pnml.NewDecoder(xmlFile)
	var p = new(pnml.Net)
	err = decoder.Build(p)
	if err != nil {
		log.Println("Error decoding PNML file:", err)
		os.Exit(1)
		return
	}

	hl, err := hlnet.Build(p)
	if err != nil {
		log.Println("Error decoding PNML file:", err)
		os.Exit(1)
		return
	}

	s := hlnet.NewStepper(hl)

	if *flagstat {
		elapsed := time.Since(start)
		fmt.Fprintf(os.Stdout, "# net %s, %d place(s), %d transition(s), %.3fs\n", hl.Name, len(hl.Places), len(hl.Trans), elapsed.Seconds())
		fmt.Fprintf(os.Stdout, "%s\n", hl)
		fmt.Fprintf(os.Stdout, "%s\n", s)
	}

	if needformulas {
		if *flagreach {
			xmlFile, err = os.Open(filepath.Join(dirname, "ReachabilityCardinality.xml"))
		}
		if *flagfire {
			xmlFile, err = os.Open(filepath.Join(dirname, "ReachabilityFireability.xml"))
		}

		if err != nil {
			log.Fatal("Error opening file:", err)
			os.Exit(1)
			return
		}
		defer xmlFile.Close()
		decoder := formula.NewDecoder(xmlFile)
		queries, err := decoder.Build()
		if err != nil {
			log.Fatal("Error decoding Formula file:", err)
			os.Exit(1)
			return
		}
		for _, q := range queries {
			v := hlnet.Evaluate(q, s.Marking)
			if !*flaghideundef || (*flaghideundef && (v != hlnet.UNDEF)) {
				if !*flaghidetrivial || (*flaghidetrivial && !q.IsTrivial()) {
					if *flagshowqueries {
						fmt.Fprint(os.Stdout, q.String())
						fmt.Fprintf(os.Stdout, "VERDICT\t%s\n", v)
						fmt.Println("----------------------------------")
					} else {
						fmt.Fprintf(os.Stdout, "FORMULA %s %s\n", q.ID, v)
					}
				}
			}
		}
	}
}
