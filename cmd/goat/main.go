package main

import (
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/tracking"
)

func main() {
	projectPath := flag.String("p", "", "project path")
	stableBranch := flag.String("s", "", "stable branch")
	publishBranch := flag.String("b", "", "publish branch")
	workers := flag.Int("w", 4, "number of workers")
	flag.Parse()
	if *projectPath == "" || *stableBranch == "" || *publishBranch == "" {
		log.Fatalf("project path, stable branch and publish branch are required")
	}
	var differ diff.DifferInterface
	differ, err := diff.NewDifferV2(*projectPath, *stableBranch, *publishBranch, *workers)
	if err != nil {
		log.Fatalf("failed to create differ: %v", err)
	}
	changes, err := differ.AnalyzeChanges()
	if err != nil {
		log.Fatalf("failed to analyze changes2: %v", err)
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
	// for _, change := range changes {
	// 	log.Printf("change: %v", change)
	// }
	track, err := tracking.NewIncreamentTrack(*projectPath, changes[5], nil, tracking.GranularityLine)
	if err != nil {
		log.Fatalf("failed to create track: %v", err)
	}
	n, err := track.Track()
	if err != nil {
		log.Fatalf("failed to track: %v", err)
	}
	log.Printf("tracked %d lines", n)
	fmt.Printf("%s", track.Bytes())
}
