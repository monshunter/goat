package goat

import (
	"fmt"
	"go/printer"
	"os"
	"path/filepath"
	"sort"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking/increment"
	"github.com/monshunter/goat/pkg/utils"
)

// getDiff gets the diff
func getDiff(cfg *config.Config) ([]*diff.FileChange, error) {
	var differ diff.DifferInterface
	var err error
	// if the repository is new, use the diff.NewDifferInit
	if cfg.IsNewRepository() {
		differ, err = diff.NewDifferInit(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get differ: %v", err)
		}
		return differ.AnalyzeChanges()
	}

	switch cfg.DiffPrecision {
	case 1:
		differ, err = diff.NewDifferV1(cfg)
	case 2:
		differ, err = diff.NewDifferV2(cfg)
	case 3:
		differ, err = diff.NewDifferV3(cfg)
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

// goatFile is the goat file
type goatFile struct {
	filename string
	content  string
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
func getMainPackageInfos(cfg *config.Config, projectRoot string, goModule string) ([]maininfo.MainPackageInfo, error) {
	return getMainPackageInfosWithConfig(cfg, projectRoot, goModule)
}

// getMainPackageInfosWithConfig gets the main package infos with configuration
func getMainPackageInfosWithConfig(cfg *config.Config, projectRoot string, goModule string) ([]maininfo.MainPackageInfo, error) {
	mainPkgInfo, err := maininfo.NewMainInfoWithConfig(cfg, projectRoot, goModule)
	if err != nil {
		log.Errorf("Failed to get main info: %v", err)
		return nil, err
	}
	if len(mainPkgInfo.MainPackageInfos) == 0 {
		log.Errorf("Warning: no main package info found")
		return nil, fmt.Errorf("warning: no main package info found")
	}
	return mainPkgInfo.MainPackageInfos, nil
}

// applyMainEntries applies the main entries
func applyMainEntries(cfg *config.Config, goModule string,
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
		codes := increment.GetMainEntryInsertData(cfg.GoatPackageAlias, i)
		_, err := mainInfo.ApplyMainEntry(cfg.PrinterConfig(), cfg.GoatPackageAlias, importPath, codes)
		if err != nil {
			log.Errorf("Failed to apply main entry: %v", err)
			return err
		}
	}
	return nil
}

func handleGoatDelete(cfg *printer.Config, fileContents string, goatImportPath string, goatPackageAlias string) (int, string, error) {
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
			bytes, err := utils.DeleteImport(cfg, goatImportPath, goatPackageAlias, "", []byte(content))
			if err != nil {
				return 0, "", err
			}
			return count, string(bytes), nil
		}
		return count, content, nil
	}
	return count, fileContents, nil
}

// handleGoatInsert handles the goat insert
func handleGoatInsert(cfg *printer.Config, fileContents string, goatImportPath string, goatPackageAlias string) (int, string, error) {
	count, content, err := utils.ReplaceWithRegexp(config.TrackInsertRegexp, fileContents,
		func(older string) (newer string) {
			return increment.GetPackageInsertDataString()
		})
	if err != nil {
		log.Errorf("Failed to handle goat insert: %v", err)
		return 0, "", err
	}
	if count > 0 {
		// add the import path
		bytes, err := utils.AddImport(cfg, goatImportPath, goatPackageAlias, "", []byte(content))
		if err != nil {
			log.Errorf("Failed to add import: %v", err)
			return 0, "", err
		}
		return count, string(bytes), nil
	}
	return count, fileContents, nil
}

// resetGoatGenerate resets the goat generate
func resetGoatGenerate(fileContents string) (int, string, error) {
	return utils.ReplaceWithRegexp(config.TrackGenerateEndRegexp, fileContents,
		func(older string) (newer string) {
			return increment.GetPackageInsertDataString()
		})
}

// resetGoatMain resets the goat main
func resetGoatMain(cfg *printer.Config, fileContents string, goatImportPath string, goatPackageAlias string) (int, string, error) {
	count, content, err := utils.ReplaceWithRegexp(config.TrackMainEntryEndRegexp, fileContents,
		func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("Failed to reset goat main: %v", err)
		return 0, "", err
	}
	if count > 0 {
		// handle if there is +goat:generate
		indicates := config.TrackGenerateEndRegexp.FindAllStringIndex(content, -1)
		if len(indicates) == 0 {
			// delete the import path
			bytes, err := utils.DeleteImport(cfg, goatImportPath, goatPackageAlias, "", []byte(content))
			if err != nil {
				log.Errorf("Failed to delete import: %v", err)
				return 0, "", err
			}
			return count, string(bytes), nil
		}
	}
	return count, content, nil
}

func prepareFiles(cfg *config.Config) (files []string, err error) {
	files = make([]string, 0)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		log.Debugf("Prepare files: %s", path)
		if err != nil {
			log.Errorf("Failed to walk: %v", err)
			return err
		}
		if info.IsDir() {
			if !cfg.IsTargetDir(path) {
				// Log when skipping nested modules for user awareness
				if cfg.SkipNestedModules && path != "." && cfg.IsBelongNestedModule(path) {
					log.Warningf("Skipping nested module directory: %s", path)
				}
				return filepath.SkipDir
			}
			return nil
		}
		if !utils.IsGoFile(path) {
			return nil
		}
		// skip goat_generated.go
		if path == cfg.GoatGeneratedFile() {
			return nil
		}
		// get relative path
		files = append(files, utils.Rel(".", path))
		return nil
	})
	return files, err
}
