package goat

import (
	"os"
	"sync"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/utils"
)

type CleanExecutor struct {
	cfg              *config.Config
	goatImportPath   string
	goatPackageAlias string
	files            []goatFile
}

func NewCleanExecutor(cfg *config.Config) *CleanExecutor {
	executor := &CleanExecutor{
		cfg:   cfg,
		files: make([]goatFile, 0),
	}
	executor.goatImportPath = utils.GoatPackageImportPath(config.GoModuleName(), cfg.GoatPackagePath)
	executor.goatPackageAlias = cfg.GoatPackageAlias
	return executor
}

func (e *CleanExecutor) Run() error {
	log.Infof("Cleaning project")
	if err := e.prepare(); err != nil {
		log.Errorf("failed to prepare: %v", err)
		return err
	}
	if err := e.clean(); err != nil {
		log.Errorf("failed to clean: %v", err)
		return err
	}
	log.Infof("Cleaned project")
	return nil
}

func (e *CleanExecutor) prepare() error {
	log.Infof("Preparing files")
	var err error
	files, err := prepareFiles(e.cfg)
	if err != nil {
		log.Errorf("failed to prepare files: %v", err)
		return err
	}
	if err := e.prepareContents(files); err != nil {
		log.Errorf("failed to prepare contents: %v", err)
		return err
	}
	log.Infof("Prepared %d files", len(e.files))
	return nil
}

func (e *CleanExecutor) prepareContents(files []string) error {
	if e.cfg.Threads == 1 {
		return e.prepareContentsSequential(files)
	}
	return e.prepareContentsParallel(files)
}

func (e *CleanExecutor) prepareContentsSequential(files []string) error {
	for _, file := range files {
		content, changed, err := e.prepareContent(file)
		if err != nil {
			log.Errorf("failed to prepare content: %v", err)
			return err
		}
		if changed {
			e.files = append(e.files, goatFile{
				filename: file,
				content:  content,
			})
		}
	}
	return nil
}

// prepareContentsParallel
// prepareContentsParallel is the parallel version of prepareContents
// It processes files concurrently using a worker pool pattern to limit goroutine count
// Algorithm:
// 1. Uses a semaphore channel to limit concurrent goroutines (e.g. e.cfg.Threads)
// 2. Each worker processes a file and:
//   - Reads and processes file content (prepareContent)
//   - Stores result in thread-safe slice (goatFiles)
//   - Reports errors via channel
//
// Complexity:
// - Time: O(n) where n is number of files (parallelism reduces constant factor)
// - Space: O(n) for storing results
// Correctness:
// - WaitGroup ensures all workers complete
// - Error channel provides immediate error propagation
// - Slice indexing prevents race conditions on results
// Optimizations:
// 1. Could batch files to reduce goroutine overhead
// 2. Could use sync.Pool for temporary buffers
// 3. Could implement work stealing for better load balancing
// 4. Could add context cancellation support
func (e *CleanExecutor) prepareContentsParallel(files []string) error {
	var wg sync.WaitGroup
	count := len(files)
	wg.Add(count)

	sem := make(chan struct{}, e.cfg.Threads)
	errChan := make(chan error, count)

	// We'll collect potential files here first
	goatFilesChan := make(chan goatFile, count)

	// Launch all goroutines first
	for i, file := range files {
		sem <- struct{}{}
		go func(i int, file string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			content, changed, err := e.prepareContent(file)
			if err != nil {
				log.Errorf("failed to prepare content: %v", err)
				errChan <- err
				return
			}

			if changed {
				goatFilesChan <- goatFile{
					filename: file,
					content:  content,
				}
			}

			errChan <- nil
		}(i, file)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(goatFilesChan)
		close(errChan)
	}()

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Collect results (only changed files)
	for file := range goatFilesChan {
		e.files = append(e.files, file)
	}

	return nil
}

func (e *CleanExecutor) prepareContent(filename string) (string, bool, error) {
	log.Debugf("preparing content for file: %s", filename)
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Errorf("failed to read file: %v", err)
		return "", false, err
	}
	content := string(contentBytes)
	changed := false
	// handle +goat:delete
	log.Debugf("replacing +goat:delete for file: %s", filename)
	count, newContent, err := utils.ReplaceWithRegexp(config.TrackDeleteEndRegexp,
		content, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("failed to replace +goat:delete: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:insert
	log.Debugf("replacing +goat:insert for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackInsertRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("failed to replace +goat:insert: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:generate
	log.Debugf("replacing +goat:generate for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackGenerateEndRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("failed to replace +goat:generate: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:main
	log.Debugf("replacing +goat:main for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackMainEntryEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		log.Errorf("failed to replace +goat:main: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:user
	log.Debugf("replacing +goat:user for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackUserEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		log.Errorf("failed to replace +goat:user: %v", err)
		return "", false, err
	}
	changed = changed || count > 0

	if changed {
		// remove import
		log.Debugf("deleting import for file: %s", filename)
		bytes, err := utils.DeleteImport(e.cfg.PrinterConfig(), e.goatImportPath, e.goatPackageAlias, "", []byte(newContent))
		if err != nil {
			log.Errorf("failed to delete import: %v", err)
			return "", false, err
		}
		return string(bytes), true, nil
	}

	return newContent, changed, nil
}

func (e *CleanExecutor) clean() error {
	log.Infof("Cleaning contents")
	var err error
	if e.cfg.Threads == 1 {
		err = e.cleanContentsSequential()
	} else {
		err = e.cleanContentsParallel()
	}
	if err != nil {
		log.Errorf("failed to clean contents: %v", err)
		return err
	}

	log.Infof("Total cleaned files: %d", len(e.files))
	log.Debugf("Removing goat generated file: %s", e.cfg.GoatGeneratedFile())
	os.Remove(e.cfg.GoatGeneratedFile())
	// remove goat package if empty
	log.Debugf("Checking if goat package is empty: %s", e.cfg.GoatPackagePath)
	empty, err := utils.IsDirEmpty(e.cfg.GoatPackagePath)
	if err != nil {
		log.Errorf("failed to check if goat package is empty: %v", err)
		return err
	}
	if empty {
		log.Debugf("Removing goat package: %s", e.cfg.GoatPackagePath)
		os.RemoveAll(e.cfg.GoatPackagePath)
	}
	return nil
}

func (e *CleanExecutor) cleanContentsSequential() error {
	for _, file := range e.files {
		log.Debugf("Cleaning file: %s", file.filename)
		err := utils.FormatAndSave(file.filename, []byte(file.content), e.cfg.PrinterConfig())
		if err != nil {
			log.Errorf("Failed to format and save file: %v", err)
			return err
		}
	}
	return nil
}

// cleanContentsParallel
// cleanContentsParallel is the parallel version of cleanContentsSequential
// It processes files concurrently using a worker pool pattern to limit goroutine count
// Algorithm:
// 1. Uses a semaphore channel to limit concurrent goroutines (e.g. e.cfg.Threads)
// 2. Each worker:
//   - Reads file stats (permissions)
//   - Writes cleaned content with original permissions
//   - Reports errors via channel
//
// Complexity:
// - Time: O(n) where n is number of files (parallelism reduces constant factor)
// - Space: O(n) for error channel and semaphore
// Correctness:
// - WaitGroup ensures all workers complete
// - Error channel provides immediate error propagation
// - Original permissions preserved
// Optimizations:
// 1. Could batch files to reduce goroutine overhead
// 2. Could use sync.Pool for temporary buffers
// 3. Could implement work stealing for better load balancing
// 4. Could add context cancellation support
// 5. Could pre-allocate error channel capacity
func (e *CleanExecutor) cleanContentsParallel() error {
	var wg sync.WaitGroup
	count := len(e.files)
	wg.Add(count)
	sem := make(chan struct{}, e.cfg.Threads)
	errChan := make(chan error, count)
	for _, file := range e.files {
		sem <- struct{}{}
		go func(file goatFile) {
			defer func() {
				<-sem
				wg.Done()
			}()
			log.Debugf("Cleaning file: %s", file.filename)
			err := utils.FormatAndSave(file.filename, []byte(file.content), e.cfg.PrinterConfig())
			if err != nil {
				log.Errorf("Failed to format and save file: %v", err)
				errChan <- err
				return
			}
		}(file)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
