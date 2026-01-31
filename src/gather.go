package manalyzer

import (
	"errors"
	"fmt"
	"log"
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

func GatherAllDemos() ([]*api.Match, error) {
	rgx := "*.dem"
	hits, err := filepath.Glob(rgx)
	if err != nil {
		log.Printf("Filepath error %v", err)
		return nil, err
	}
	if len(hits) == 0 {
		log.Printf("No .dem files found")
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
