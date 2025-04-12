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
	for _, change := range changes {
		log.Printf("change: %v", change)
	}
	change := changes[1]
	var track tracking.Tracker
	track, err = tracking.NewIncreamentTrack(*projectPath, change, nil, tracking.GranularityLine)
	if err != nil {
		log.Fatalf("failed to create track:%s %v", change.Path, err)
	}
	n, err := track.Track()
	if err != nil {
		log.Fatalf("failed to track:%s %v", change.Path, err)
	}
	count, err := track.Replace("TRACK_ID", tracking.IncrementReplace(10))
	if err != nil {
		log.Fatalf("failed to calibrate:%s %v", change.Path, err)
	}
	log.Printf("tracked %d lines, replaced %d times", n, count)
	fmt.Printf("%s", track.Bytes())
}
