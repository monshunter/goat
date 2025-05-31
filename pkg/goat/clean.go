package goat

import (
	"os"
	"sync"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/utils"
)

// CleanExecutor is the executor for the clean
type CleanExecutor struct {
	cfg              *config.Config
	goatImportPath   string
	goatPackageAlias string
	files            []goatFile
}

// NewCleanExecutor creates a new clean executor
func NewCleanExecutor(cfg *config.Config) *CleanExecutor {
	executor := &CleanExecutor{
		cfg:   cfg,
		files: make([]goatFile, 0),
	}
	executor.goatImportPath = utils.GoatPackageImportPath(config.GoModuleName(), cfg.GoatPackagePath)
	executor.goatPackageAlias = cfg.GoatPackageAlias
	return executor
}

// Run runs the clean executor
func (c *CleanExecutor) Run() error {
	log.Infof("Cleaning project")
	if err := c.prepare(); err != nil {
		log.Errorf("Failed to prepare: %v", err)
		return err
	}
	if err := c.clean(); err != nil {
		log.Errorf("Failed to clean: %v", err)
		return err
	}
	log.Infof("Cleaned project")
	return nil
}

// prepare prepares the files
func (c *CleanExecutor) prepare() error {
	log.Infof("Preparing files")
	var err error
	files, err := prepareFiles(c.cfg)
	if err != nil {
		log.Errorf("Failed to prepare files: %v", err)
		return err
	}
	if err := c.prepareContents(files); err != nil {
		log.Errorf("Failed to prepare contents: %v", err)
		return err
	}
	log.Infof("Prepared %d files", len(c.files))
	return nil
}

// prepareContents prepares the contents of the files
func (c *CleanExecutor) prepareContents(files []string) error {
	if c.cfg.Threads == 1 {
		return c.prepareContentsSequential(files)
	}
	return c.prepareContentsParallel(files)
}

// prepareContentsSequential prepares the contents of the files sequentially
func (c *CleanExecutor) prepareContentsSequential(files []string) error {
	for _, file := range files {
		content, changed, err := c.prepareContent(file)
		if err != nil {
			log.Errorf("Failed to prepare content: %v", err)
			return err
		}
		if changed {
			c.files = append(c.files, goatFile{
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
// 1. Uses a semaphore channel to limit concurrent goroutines (c.g. c.cfg.Threads)
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
func (c *CleanExecutor) prepareContentsParallel(files []string) error {
	var wg sync.WaitGroup
	count := len(files)
	wg.Add(count)

	sem := make(chan struct{}, c.cfg.Threads)
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

			content, changed, err := c.prepareContent(file)
			if err != nil {
				log.Errorf("Failed to prepare content: %v", err)
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
		c.files = append(c.files, file)
	}

	return nil
}

func (c *CleanExecutor) prepareContent(filename string) (string, bool, error) {
	log.Debugf("Preparing content for file: %s", filename)
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Errorf("Failed to read file: %v", err)
		return "", false, err
	}
	content := string(contentBytes)
	changed := false
	// handle +goat:delete
	log.Debugf("Replacing +goat:delete for file: %s", filename)
	count, newContent, err := utils.ReplaceWithRegexp(config.TrackDeleteEndRegexp,
		content, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("Failed to replace +goat:delete: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:insert
	log.Debugf("Replacing +goat:insert for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackInsertRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("Failed to replace +goat:insert: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:generate
	log.Debugf("Replacing +goat:generate for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackGenerateEndRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("Failed to replace +goat:generate: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:main
	log.Debugf("Replacing +goat:main for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackMainEntryEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		log.Errorf("Failed to replace +goat:main: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:user
	log.Debugf("Replacing +goat:user for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackUserEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		log.Errorf("Failed to replace +goat:user: %v", err)
		return "", false, err
	}
	changed = changed || count > 0

	if changed {
		// remove import
		log.Debugf("Deleting import for file: %s", filename)
		bytes, err := utils.DeleteImport(c.cfg.PrinterConfig(), c.goatImportPath, c.goatPackageAlias, "", []byte(newContent))
		if err != nil {
			log.Errorf("Failed to delete import: %v", err)
			return "", false, err
		}
		return string(bytes), true, nil
	}

	return newContent, changed, nil
}

// clean cleans the contents of the files
func (c *CleanExecutor) clean() error {
	log.Infof("Cleaning contents")
	var err error
	if c.cfg.Threads == 1 {
		err = c.cleanContentsSequential()
	} else {
		err = c.cleanContentsParallel()
	}
	if err != nil {
		log.Errorf("Failed to clean contents: %v", err)
		return err
	}

	log.Infof("Total cleaned files: %d", len(c.files))
	log.Debugf("Removing goat generated file: %s", c.cfg.GoatGeneratedFile())
	os.Remove(c.cfg.GoatGeneratedFile())
	// remove goat package if empty
	log.Debugf("Checking if goat package is empty: %s", c.cfg.GoatPackagePath)
	empty, err := utils.IsDirEmpty(c.cfg.GoatPackagePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("Goat package does not exist: %s", c.cfg.GoatPackagePath)
			return nil
		}
		log.Errorf("Failed to check if goat package is empty: %v", err)
		return err
	}
	if empty {
		log.Debugf("Removing goat package: %s", c.cfg.GoatPackagePath)
		os.RemoveAll(c.cfg.GoatPackagePath)
	}
	return nil
}

// cleanContentsSequential cleans the contents of the files sequentially
func (c *CleanExecutor) cleanContentsSequential() error {
	for _, file := range c.files {
		log.Debugf("Cleaning file: %s", file.filename)
		err := utils.FormatAndSave(file.filename, []byte(file.content), c.cfg.PrinterConfig())
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
// 1. Uses a semaphore channel to limit concurrent goroutines (c.g. c.cfg.Threads)
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
func (c *CleanExecutor) cleanContentsParallel() error {
	var wg sync.WaitGroup
	count := len(c.files)
	wg.Add(count)
	sem := make(chan struct{}, c.cfg.Threads)
	errChan := make(chan error, count)
	for _, file := range c.files {
		sem <- struct{}{}
		go func(file goatFile) {
			defer func() {
				<-sem
				wg.Done()
			}()
			log.Debugf("Cleaning file: %s", file.filename)
			err := utils.FormatAndSave(file.filename, []byte(file.content), c.cfg.PrinterConfig())
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
