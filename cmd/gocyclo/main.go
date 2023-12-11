// Copyright 2013 Frederik Zipp. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Gocyclo calculates the cyclomatic complexities of functions and
// methods in Go source code.
//
// Usage:
//
//	gocyclo [<flag> ...] <Go file or directory> ...
//
// Flags:
//
//	-over N               show functions with complexity > N only and
//	                      return exit code 1 if the output is non-empty
//	-top N                show the top N most complex functions only
//	-avg, -avg-short      show the average complexity;
//	                      the short option prints the value without a label
//	-ignore REGEX         exclude files matching the given regular expression
//
// The output fields for each line are:
// <complexity> <package> <function> <file:line:column>
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fzipp/gocyclo"
)

const usageDoc = `Calculate cyclomatic complexities of Go functions.
Usage:
    gocyclo [flags] <Go file or directory> ...

Flags:
    -over N               show functions with complexity > N only and
                          return exit code 1 if the set is non-empty
    -top N                show the top N most complex functions only
    -avg, -avg-short      show the average complexity over all functions;
                          the short option prints the value without a label
    -ignore REGEX         exclude files matching the given regular expression

The output fields for each line are:
<complexity> <package> <function> <file:line:column>
`

func main() {
	over := flag.Int("over", 0, "show functions with complexity > N only")
	under := flag.Int("under", 0, "show functions with complexity < N only")
	top := flag.Int("top", -1, "show the top N most complex functions only")
	avg := flag.Bool("avg", false, "show the average complexity")
	avgShort := flag.Bool("avg-short", false, "show the average complexity without a label")
	ignore := flag.String("ignore", "", "exclude files matching the given regular expression")
	report := flag.String("report", "", "show the general report")

	log.SetFlags(0)
	log.SetPrefix("gocyclo: ")
	flag.Usage = usage
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		usage()
	}

	var breakPoints []int
	if *report != "" {
		stringBreakPoints := strings.Split(*report, ",")
		breakPoints = make([]int, len(stringBreakPoints))
		for _, breakPoint := range stringBreakPoints {
			point, err := strconv.Atoi(breakPoint)
			if err != nil {
				usage()
			}
			breakPoints = append(breakPoints, point)
		}
	}

	allStats := gocyclo.Analyze(paths, regex(*ignore))
	shownStats := allStats.SortAndFilter(*top, *over, *under)

	printStats(shownStats)
	if *avg || *avgShort {
		printAverage(allStats, *avgShort)
	}

	if *report != "" {
		printReport(shownStats, breakPoints)
	}

	if *over > 0 && len(shownStats) > 0 {
		os.Exit(1)
	}
}

func regex(expr string) *regexp.Regexp {
	if expr == "" {
		return nil
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		log.Fatal(err)
	}
	return re
}

func printStats(s gocyclo.Stats) {
	for _, stat := range s {
		fmt.Println(stat)
	}
}

func printAverage(s gocyclo.Stats, short bool) {
	if !short {
		fmt.Print("Average: ")
	}
	fmt.Printf("%.3g\n", s.AverageComplexity())
}

func printReport(s gocyclo.Stats, breakPoints []int) {
	totalComplexity := int(s.TotalComplexity())
	last := -1
	fmt.Println("RANGE\t|COUNT\t|PERCENT")
	for key, value := range s.Report(breakPoints) {
		fmt.Printf("[%d, %d) - %d - %d", key, last, value, value/totalComplexity*100)
	}
}

func usage() {
	_, _ = fmt.Fprint(os.Stderr, usageDoc)
	os.Exit(2)
}
