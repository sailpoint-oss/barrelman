package barrelman

import (
	"path/filepath"
	"strings"

	navigator "github.com/sailpoint-oss/navigator"
)

func prepareLintProjects(files []string) map[string]*navigator.Project {
	if len(files) == 0 {
		return nil
	}
	workspace := navigator.NewWorkspace()
	projects := make(map[string]*navigator.Project, len(files))
	for _, file := range files {
		project, err := workspace.OpenProject(file)
		if err != nil {
			continue
		}
		projects[fileToURI(file)] = project
	}
	return projects
}

func resolveLintWorkspaceRoot(files []string, configuredRoot string) string {
	if configuredRoot != "" {
		if abs, err := filepath.Abs(configuredRoot); err == nil {
			return abs
		}
		return filepath.Clean(configuredRoot)
	}
	return inferCommonWorkspaceRoot(files)
}

func inferCommonWorkspaceRoot(files []string) string {
	if len(files) == 0 {
		return ""
	}
	common := absDir(files[0])
	for _, file := range files[1:] {
		common = commonParentDir(common, absDir(file))
	}
	return common
}

func absDir(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return filepath.Dir(filepath.Clean(abs))
}

func commonParentDir(a, b string) string {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	for a != "" && a != "." {
		if sameOrParentDir(a, b) {
			return a
		}
		parent := filepath.Dir(a)
		if parent == a {
			return a
		}
		a = parent
	}
	return a
}

func sameOrParentDir(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || rel == "" || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}
