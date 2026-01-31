package manalyzer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/akiver/cs-demo-analyzer/pkg/api"
	"github.com/akiver/cs-demo-analyzer/pkg/api/constants"
)

var ErrNoDemos = errors.New("no .dem files found")

func GatherDemo(demoPath string) (*api.Match, error) {
	match, err := api.AnalyzeDemo(demoPath, api.AnalyzeDemoOptions{
		IncludePositions: false,
		Source:           constants.DemoSourceValve,
	})

	if err != nil {
		return nil, err
	}

	return match, nil
}

// GatherAllDemosFromPath recursively finds and analyzes all .dem files in a directory tree
func GatherAllDemosFromPath(basePath string) ([]*api.Match, error) {
	var matches []*api.Match
	var errs []error
	var demoCount int

	// Validate base path exists
	info, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("base path does not exist: %s", basePath)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot access base path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base path is not a directory: %s", basePath)
	}

	// Walk directory tree recursively
	err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Continue with other directories on error
			return nil
		}

		if d.IsDir() {
			return nil // Skip directories
		}

		if filepath.Ext(path) != ".dem" {
			return nil // Skip non-.dem files
		}

		demoCount++

		// Analyze demo
		match, err := GatherDemo(path)
		if err != nil {
			errMsg := fmt.Errorf("failed to analyze %s: %w", path, err)
			errs = append(errs, errMsg)
			return nil // Continue processing other files
		}

		matches = append(matches, match)

		return nil
	})

	if err != nil {
		errs = append(errs, fmt.Errorf("directory walk error: %w", err))
	}

	// Return appropriate error if no demos found
	if demoCount == 0 {
		return nil, ErrNoDemos
	}

	if len(matches) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all %d demos failed to parse: %w", demoCount, errors.Join(errs...))
	}

	// Return partial results with errors
	if len(errs) > 0 {
		return matches, errors.Join(errs...)
	}

	return matches, nil
}

func GatherAllDemos() ([]*api.Match, error) {
	rgx := "*.dem"
	hits, err := filepath.Glob(rgx)
	if err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return nil, ErrNoDemos
	}

	var matches []*api.Match
	var errs []error
	for _, path := range hits {
		match, err := GatherDemo(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			continue
		}
		matches = append(matches, match)
	}

	if len(errs) > 0 {
		return matches, errors.Join(errs...)
	}

	return matches, nil

}
