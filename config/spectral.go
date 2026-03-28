package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sailpoint-oss/barrelman/rulesets"
)

// SpectralFiles lists the config filenames searched for Spectral rulesets, in
// priority order. Barrelman-owned locations come first; legacy Telescope
// locations remain supported after that.
var SpectralFiles = []string{
	".barrelman/spectral.yaml",
	".barrelman/spectral.yml",
	".barrelman/spectral.json",
	".telescope/spectral.yaml",
	".telescope/spectral.yml",
	".telescope/spectral.json",
	".spectral.yaml",
	".spectral.yml",
	".spectral.json",
}

// FindSpectralRuleset searches the workspace root for a Spectral ruleset file,
// returning the full path to the first match. Returns ("", nil) when no file
// is found.
func FindSpectralRuleset(workspaceRoot string) (string, error) {
	for _, name := range SpectralFiles {
		full := filepath.Join(workspaceRoot, name)
		info, err := os.Stat(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("stat %s: %w", full, err)
		}
		if !info.IsDir() {
			return full, nil
		}
	}
	return "", nil
}

// LoadSpectralRuleset discovers and loads the first available Spectral ruleset
// file in the workspace. Returns (nil, nil) when no file is found.
func LoadSpectralRuleset(workspaceRoot string) (*rulesets.RuleSet, error) {
	path, err := FindSpectralRuleset(workspaceRoot)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	return rulesets.LoadFile(path)
}
