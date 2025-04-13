package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking"
	"github.com/monshunter/goat/pkg/tracking/increament"
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

	mainInfo, err := maininfo.NewMainInfo(*projectPath)
	if err != nil {
		log.Fatalf("failed to create main info: %v", err)
	}
	result, err := json.MarshalIndent(mainInfo, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal main info: %v", err)
	}
	_ = result
	// log.Printf("main info: %s", string(result))
	// return
	change := changes[1]
	var track tracking.Tracker
	track, err = tracking.NewIncreamentTrack(*projectPath, change, nil, config.GranularityLine)
	if err != nil {
		log.Fatalf("failed to create track:%s %v", change.Path, err)
	}
	n, err := track.Track()
	if err != nil {
		log.Fatalf("failed to track:%s %v", change.Path, err)
	}
	start := 10
	_, err = track.Replace(`"github.com/monshunter/goat"`, tracking.IncreamentReplaceImport("goat", "github.com/monshunter/goat"))
	count, err := track.Replace(`goat.Track(TRACK_ID)`, tracking.IncreamentReplaceStmt("goat", start))
	if err != nil {
		log.Fatalf("failed to calibrate:%s %v", change.Path, err)
	}
	log.Printf("tracked %d lines, replaced %d times", n, count)
	fmt.Printf("%s", track.Bytes())
	values := increament.NewValues("goat", "1.0.0", "goat", false)
	idx := make([]int, count)
	for i := 0; i < count; i++ {
		idx[i] = i + start
	}
	values.AddTrackIds(idx)
	values.AddComponent(1, "LoginComponent", idx[:len(idx)/2])
	values.AddComponent(2, "DashboardComponent", idx[len(idx)/2:])
	values.Save("tmp/track.go")
}
