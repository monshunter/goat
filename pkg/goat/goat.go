package goat

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
)

// debugChanges debugs the changes
func debugChanges(changes []*diff.FileChange) {
	result, err := json.MarshalIndent(changes, "", "  ")
	if err != nil {
		log.Printf("failed to marshal changes: %v", err)
		return
	}
	log.Printf("changes: %s", string(result))
}

// debugMainInfo debugs the main info
func debugMainInfo(mainPkgInfo *maininfo.MainInfo) {
	result, err := json.MarshalIndent(mainPkgInfo, "", "  ")
	if err != nil {
		log.Printf("failed to marshal main info: %v", err)
		return
	}
	log.Printf("main info: %s", string(result))
}

// getDiff gets the diff
func getDiff(cfg *config.Config) ([]*diff.FileChange, error) {
	var differ diff.DifferInterface
	var err error
	switch cfg.DiffPrecision {
	case 1:
		differ, err = diff.NewDiffer(cfg.ProjectRoot, cfg.StableBranch, cfg.PublishBranch, cfg.DiffPrecision)
	case 2:
		differ, err = diff.NewDifferV2(cfg.ProjectRoot, cfg.StableBranch, cfg.PublishBranch, cfg.DiffPrecision)
	case 3:
		differ, err = diff.NewDifferV3(cfg.ProjectRoot, cfg.StableBranch, cfg.PublishBranch, cfg.DiffPrecision)
	case 4:
		differ, err = diff.NewDifferV4(cfg.ProjectRoot, cfg.StableBranch, cfg.PublishBranch, cfg.DiffPrecision)
	default:
		return nil, fmt.Errorf("invalid diff precision: %d", cfg.DiffPrecision)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get differ: %v", err)
	}
	return differ.AnalyzeChanges()
}

// trackIdxInterval is the interval of the track
type trackIdxInterval struct {
	start int
	end   int
}

// componentTrackIdx is the index of the component track
type componentTrackIdx struct {
	componentId int
	component   string
	trackIdx    []int
}

func debugComponentTrackIdxs(componentTrackIdxs []componentTrackIdx) {
	for _, component := range componentTrackIdxs {
		log.Printf("component: %d, %s, %d, %v\n",
			component.componentId, component.component, len(component.trackIdx), component.trackIdx)
	}
}
