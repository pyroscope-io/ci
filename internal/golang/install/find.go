package install

import (
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	InstallationFileName = "pyroscope_test.go"
)

type InstallationCandidate struct {
	Path        string
	PackageName string
}

// FindCandidateFiles finds files candidate for installation
func FindCandidateFiles(basePath string) (map[string]InstallationCandidate, error) {
	// Use a map to not have any duplicates
	testPackages := make(map[string]InstallationCandidate)

	// Find all test files
	err := filepath.Walk(basePath,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if isTestFile(p) {
				dirName, _ := filepath.Split(p)

				if _, ok := testPackages[dirName]; ok {
					return nil
				}

				// TODO: this is costly, we should optimize to do it just once per package
				packageName, err := getPackageNameFromFile(p)

				if err != nil {
					return err
				}

				testPackages[dirName] = InstallationCandidate{
					Path:        path.Join(dirName, InstallationFileName),
					PackageName: packageName,
				}
			}

			return nil
		})

	return testPackages, err
}

func getPackageNameFromFile(fileName string) (string, error) {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, fileName, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}

	return ast.Name.Name, nil
}

func isTestFile(p string) bool {
	// TODO: is this enough?
	return strings.HasSuffix(p, "_test.go")
}
