package goat

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sort"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking/increament"
	"github.com/monshunter/goat/pkg/utils"
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
		differ, err = diff.NewDiffer(cfg)
	case 2:
		differ, err = diff.NewDifferV2(cfg)
	case 3:
		differ, err = diff.NewDifferV3(cfg)
	case 4:
		differ, err = diff.NewDifferV4(cfg)
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

// getComponentTrackIdxs gets the component track idxs
// fileTrackIdStartMap is the map of the file to the track idxs
// mainPackageInfos is the main package infos
func getComponentTrackIdxs(fileTrackIdStartMap map[string]trackIdxInterval, mainPackageInfos []maininfo.MainPackageInfo) []componentTrackIdx {
	// packageTrackIdxMap: package -> trackIdxs
	packageTrackIdxMap := make(map[string][]int)
	for path, interval := range fileTrackIdStartMap {
		pkg := filepath.Dir(path)
		ids := make([]int, 0, interval.end-interval.start+1)
		for i := interval.start; i <= interval.end; i++ {
			ids = append(ids, i)
		}
		packageTrackIdxMap[pkg] = append(packageTrackIdxMap[pkg], ids...)
	}

	// componentTrackIdxs: componentId -> component -> trackIdxs
	componentTrackIdxs := make([]componentTrackIdx, 0)
	for i, mainInfo := range mainPackageInfos {
		trackIdxs := make([]int, 0)
		for _, pkg := range mainInfo.Imports {
			ids, ok := packageTrackIdxMap[pkg]
			if !ok {
				continue
			}
			trackIdxs = append(trackIdxs, ids...)
		}
		sort.Ints(trackIdxs)
		component := componentTrackIdx{
			componentId: i,
			component:   mainInfo.MainDir,
			trackIdx:    trackIdxs,
		}
		componentTrackIdxs = append(componentTrackIdxs, component)
	}
	return componentTrackIdxs
}

// getTotalTrackIdxs gets the total track idxs
// fileTrackIdStartMap is the map of the file to the track idxs
func getTotalTrackIdxs(fileTrackIdStartMap map[string]trackIdxInterval) []int {
	idxs := make([]int, 0)
	for _, interval := range fileTrackIdStartMap {
		for i := interval.start; i <= interval.end; i++ {
			idxs = append(idxs, i)
		}
	}

	if len(idxs) == 0 {
		return []int{}
	}

	sort.Ints(idxs)
	slow, fast := 0, 0
	for fast < len(idxs) {
		if idxs[slow] == idxs[fast] {
			fast++
		} else {
			slow++
			idxs[slow] = idxs[fast]
		}
	}
	return idxs[:slow+1]
}

// getMainPackageInfos gets the main package infos
func getMainPackageInfos(projectRoot string, goModule string) ([]maininfo.MainPackageInfo, error) {
	mainPkgInfo, err := maininfo.NewMainInfo(projectRoot, goModule)
	if err != nil {
		log.Printf("failed to get main info: %v", err)
		return nil, err
	}
	if len(mainPkgInfo.MainPackageInfos) == 0 {
		log.Printf("warning: no main package info found")
		return nil, fmt.Errorf("warning: no main package info found")
	}
	return mainPkgInfo.MainPackageInfos, nil
}

func applyMainEntry(cfg *config.Config, goModule string,
	mainPackageInfos []maininfo.MainPackageInfo,
	componentTrackIdxs []componentTrackIdx) error {
	importPath := filepath.Join(goModule, cfg.GoatPackagePath)
	for i, mainInfo := range mainPackageInfos {
		if !cfg.IsMainEntry(mainInfo.MainDir) {
			continue
		}

		trackIdxs := componentTrackIdxs[i].trackIdx
		if len(trackIdxs) == 0 {
			continue
		}
		codes := increament.GetMainEntryInsertData(cfg.GoatPackageAlias, i)
		_, err := mainInfo.ApplyMainEntry(cfg.GoatPackageAlias, importPath, codes)
		if err != nil {
			log.Printf("failed to apply main entry: %v", err)
			return err
		}
	}
	return nil
}

func handleGoatDelete(fileContents string, goatImportPath string, goatPackageAlias string) (int, string, error) {
	count, content, err := utils.ReplaceWithRegexp(config.TrackDeleteEndRegexp, fileContents,
		func(older string) (newer string) {
			return ""
		})
	if err != nil {
		return 0, "", err
	}
	if count > 0 {
		// handle if there is +goat:generate
		indicates := config.TrackGenerateEndRegexp.FindAllStringIndex(content, -1)
		if len(indicates) == 0 {
			// delete the import path
			bytes, err := utils.DeleteImport(goatImportPath, goatPackageAlias, "", []byte(content))
			if err != nil {
				return 0, "", err
			}
			return count, string(bytes), nil
		}
		return count, content, nil
	}
	return count, fileContents, nil
}

func handleGoatInsert(fileContents string, goatImportPath string, goatPackageAlias string) (int, string, error) {
	count, content, err := utils.ReplaceWithRegexp(config.TrackInsertRegexp, fileContents,
		func(older string) (newer string) {
			return increament.GetPackageInsertDataString()
		})
	if err != nil {
		return 0, "", err
	}
	if count > 0 {
		// add the import path
		bytes, err := utils.AddImport(goatImportPath, goatPackageAlias, "", []byte(content))
		if err != nil {
			return 0, "", err
		}
		return count, string(bytes), nil
	}
	return count, fileContents, nil
}

func resetGoatGenerate(fileContents string) (int, string, error) {
	return utils.ReplaceWithRegexp(config.TrackGenerateEndRegexp, fileContents,
		func(older string) (newer string) {
			return increament.GetPackageInsertDataString()
		})
}

func resetGoatMain(fileContents string, goatImportPath string, goatPackageAlias string) (int, string, error) {
	count, content, err := utils.ReplaceWithRegexp(config.TrackMainEntryEndRegexp, fileContents,
		func(older string) (newer string) {
			return ""
		})
	if err != nil {
		return 0, "", err
	}
	if count > 0 {
		// handle if there is +goat:generate
		indicates := config.TrackGenerateEndRegexp.FindAllStringIndex(content, -1)
		if len(indicates) == 0 {
			// delete the import path
			bytes, err := utils.DeleteImport(goatImportPath, goatPackageAlias, "", []byte(content))
			if err != nil {
				return 0, "", err
			}
			return count, string(bytes), nil
		}
	}
	return count, content, nil
}
