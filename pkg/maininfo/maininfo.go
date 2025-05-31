package maininfo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/monshunter/goat/pkg/log"
	"github.com/monshunter/goat/pkg/utils"
)

// MainPackageInfo represents information about a main package
type MainPackageInfo struct {
	MainDir  string   `json:"mainDir"`
	MainFile string   `json:"mainFile"`
	Imports  []string `json:"imports"`
}

func (m *MainPackageInfo) ApplyMainEntry(cfg *printer.Config, packageAlias string, packagePath string, codes []string) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, m.MainFile, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	position := 0
	for _, decl := range f.Decls {
		if node, ok := decl.(*ast.FuncDecl); ok && node.Name.Name == "main" && node.Recv == nil {
			position = fset.Position(node.Body.Lbrace + 2).Line
			break
		}
	}
	content, err := utils.AddCodes(cfg, fset, f, position, codes)
	if err != nil {
		return nil, err
	}
	fileInfo, err := os.Stat(m.MainFile)
	if err != nil {
		return nil, err
	}

	perm := fileInfo.Mode().Perm()
	err = os.WriteFile(m.MainFile, content, perm)
	if err != nil {
		return nil, err
	}
	content, err = utils.AddImport(cfg, packagePath, packageAlias, m.MainFile, content)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(m.MainFile, content, perm)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// MainInfo represents information about a main package
type MainInfo struct {
	ProjectRoot      string            `json:"projectRoot"`
	Module           string            `json:"module"`
	MainPackageInfos []MainPackageInfo `json:"mainPackageInfos"`
}

// NewMainInfo creates a new MainInfo instance
func NewMainInfo(projectRoot string, goModule string, ignores []string) (*MainInfo, error) {
	return NewMainInfoWithConfig(projectRoot, goModule, ignores, true) // default to skipping nested modules
}

// NewMainInfoWithConfig creates a new MainInfo instance with configuration
func NewMainInfoWithConfig(projectRoot string, goModule string, ignores []string, skipNestedModules bool) (*MainInfo, error) {
	mainInfo := &MainInfo{
		ProjectRoot: projectRoot,
		Module:      goModule,
	}
	mainPackageInfos, err := mainInfo.analyzeMainPackages(ignores, skipNestedModules)
	if err != nil {
		return nil, err
	}

	if len(mainPackageInfos) == 0 {
		return nil, fmt.Errorf("warning: no main packages found")
	}

	mainInfo.MainPackageInfos = mainPackageInfos
	return mainInfo, nil
}

// analyzeMainPackagesWithConfig analyzes all main packages with configuration
func (m *MainInfo) analyzeMainPackages(ignores []string, skipNestedModules bool) ([]MainPackageInfo, error) {
	// find all Go files
	var goFiles []string
	err := filepath.Walk(m.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if !utils.IsTargetDir(path, ignores, skipNestedModules) {
				// Log when skipping nested modules for user awareness
				if skipNestedModules && path != "." && utils.IsBelongtoNestedGoModule(path) {
					log.Warningf("Skipping nested module directory: %s", path)
				}
				return filepath.SkipDir
			}
			return nil
		}

		if !utils.IsTargetFile(path, ignores, skipNestedModules) {
			return nil
		}
		goFiles = append(goFiles, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error: failed to walk directory: %w", err)
	}

	// find all main packages
	mainPackages, err := findMainPackages(goFiles)
	if err != nil {
		return nil, fmt.Errorf("error: failed to find main packages: %w", err)
	}

	// analyze each main package
	results := make([]MainPackageInfo, 0, len(mainPackages))
	for _, mainDir := range mainPackages {
		info := m.analyzeMainImports(mainDir)
		results = append(results, info)
	}
	return results, nil
}

// getRelativeDir gets the relative directory
func getRelativeDir(root, dir string) string {
	relDir, err := filepath.Rel(root, dir)
	if err != nil {
		return ""
	}
	return relDir
}

// findMainPackages finds all main packages
func findMainPackages(goFiles []string) ([]string, error) {
	visited := make(map[string]bool)
	results := make([]string, 0, len(goFiles))
	for _, file := range goFiles {
		dir := filepath.Dir(file)
		if visited[dir] {
			continue
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		if node.Name.Name == "main" {
			results = append(results, dir)
			visited[dir] = true
		}
	}
	return results, nil
}

// findMainEntryFile finds the main entry file
func findMainEntryFile(mainDir string) string {
	entries, err := os.ReadDir(mainDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".go") && strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filepath.Join(mainDir, entry.Name()), nil, parser.ParseComments)
		if err != nil {
			continue
		}
		for _, imp := range node.Decls {
			if node, ok := imp.(*ast.FuncDecl); ok && node.Name.Name == "main" && node.Recv == nil {
				return strings.Join([]string{mainDir, entry.Name()}, "/")
			}
		}
	}
	return ""
}

// analyzeMainPackage analyzes a single main package
func (m *MainInfo) analyzeMainImports(mainDir string) MainPackageInfo {
	info := MainPackageInfo{
		MainDir:  getRelativeDir(m.ProjectRoot, mainDir),
		MainFile: getRelativeDir(m.ProjectRoot, findMainEntryFile(mainDir)),
	}
	// analyze imported packages
	importedPkgs := make(map[string]bool)
	m.collectImports(mainDir, importedPkgs)

	// filter out project internal packages
	var internalPackages []string
	for pkg := range importedPkgs {
		internalPackages = append(internalPackages, getRelativeDir(m.Module, pkg))
	}
	internalPackages = append(internalPackages, info.MainDir)
	info.Imports = internalPackages
	return info
}

// collectImports collects all imports in the directory
func (m *MainInfo) collectImports(dir string, importedPkgs map[string]bool) {
	// get all Go files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(dir, name)
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}

		// add imported packages
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if importedPkgs[importPath] {
				continue
			}
			// recursively process internal packages
			if m.isInternalPackage(importPath) {
				importedPkgs[importPath] = true
				pkgDir := m.getPackageDir(importPath)
				if pkgDir != "" && pkgDir != dir {
					m.collectImports(pkgDir, importedPkgs)
				}
			}
		}
	}
}

// isInternalPackage checks if a package is a project internal package
func (m *MainInfo) isInternalPackage(importPath string) bool {
	// if there is a module prefix, check if it starts with the module prefix
	if m.Module != "" && strings.HasPrefix(importPath, m.Module) {
		return true
	}

	// if there is no module prefix, check if the package is in the project directory
	pkgDir := m.getPackageDir(importPath)
	return pkgDir != "" && strings.HasPrefix(pkgDir, m.ProjectRoot)
}

// getPackageDir gets the directory of a package
func (m *MainInfo) getPackageDir(importPath string) string {
	if m.Module != "" && strings.HasPrefix(importPath, m.Module) {
		// remove the module prefix, get the relative path
		relPath := strings.TrimPrefix(importPath, m.Module)
		relPath = strings.TrimPrefix(relPath, "/")
		return filepath.Join(m.ProjectRoot, relPath)
	}

	// try to find the package in the GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		pkgDir := filepath.Join(gopath, "src", importPath)
		if _, err := os.Stat(pkgDir); err == nil {
			return pkgDir
		}
	}

	return ""
}
