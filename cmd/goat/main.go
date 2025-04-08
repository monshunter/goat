package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/monshunter/goat/pkg/diff"
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
	start := time.Now()
	differ, err := diff.NewDiffer(*projectPath, *stableBranch, *publishBranch, *workers)
	if err != nil {
		log.Fatalf("failed to create differ: %v", err)
	}
	fmt.Println("since", time.Since(start))
	changes1, err := differ.AnalyzeChanges()
	if err != nil {
		log.Fatalf("failed to analyze changes1: %v", err)
	}
	changes2, err := differ.AnalyzeChangesV2()
	if err != nil {
		log.Fatalf("failed to analyze changes2: %v", err)
	}
	fmt.Println("since", time.Since(start))
	// compareChanges(changes1, changes2)
	_ = changes2
	_ = changes1
	fmt.Println("changes2", len(changes2))
	fmt.Println("changes1", len(changes1))
	for _, change := range changes2 {
		log.Printf("change: %v", change)
	}
}

func compareChanges(changes1, changes2 []*diff.FileChange) {
	if len(changes1) != len(changes2) {
		fmt.Printf("changes1: %d, changes2: %d\n", len(changes1), len(changes2))
		log.Fatalf("changes1 and changes2 have different length")
	}
	sort.Slice(changes1, func(i, j int) bool {
		return changes1[i].Path < changes1[j].Path
	})
	sort.Slice(changes2, func(i, j int) bool {
		return changes2[i].Path < changes2[j].Path
	})

	for _, change := range changes1 {
		sort.Slice(change.LineChanges, func(i, j int) bool {
			return change.LineChanges[i].Start < change.LineChanges[j].Start
		})
	}

	for _, change := range changes2 {
		sort.Slice(change.LineChanges, func(i, j int) bool {
			return change.LineChanges[i].Start < change.LineChanges[j].Start
		})
	}

	for i, change := range changes1 {
		if change.Path != changes2[i].Path {
			log.Fatalf("changes1 and changes2 have different path")
		}

		if len(change.LineChanges) != len(changes2[i].LineChanges) {
			fmt.Printf("path: %s, changes1: %d, changes2: %d\n", change.Path, len(change.LineChanges), len(changes2[i].LineChanges))
			fmt.Printf("changes1: %v\n", change.LineChanges)
			fmt.Printf("changes2: %v\n", changes2[i].LineChanges)
			log.Printf("changes1 and changes2 have different line changes length")
		}

		for j, lineChange := range change.LineChanges {
			if lineChange.Start != changes2[i].LineChanges[j].Start {
				log.Printf("changes1 and changes2 have different line changes start")
			}

			if lineChange.Lines != changes2[i].LineChanges[j].Lines {
				log.Printf("changes1 and changes2 have different line changes lines")
			}
		}
	}
}
