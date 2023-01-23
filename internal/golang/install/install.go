package install

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pyroscope-io/client/pyroscope"
)

// TODO: maybe upstream should export this?
var availableProfileTypes = []pyroscope.ProfileType{
	pyroscope.ProfileCPU,
	pyroscope.ProfileInuseObjects,
	pyroscope.ProfileAllocObjects,
	pyroscope.ProfileInuseSpace,
	pyroscope.ProfileAllocSpace,
	pyroscope.ProfileGoroutines,
	pyroscope.ProfileMutexCount,
	pyroscope.ProfileMutexDuration,
	pyroscope.ProfileBlockCount,
	pyroscope.ProfileBlockDuration,
}

// Install Install the pyroscope agent into test packages
// It does that by recursively finding packages with tests
// Then it generates a `pyroscope_test.go` file for each package
// With the profile type and app name specified
func Install(basePath string, appName string, profileTypes []string) error {
	type TestPackage struct {
		Path        string
		PackageName string
	}

	// Use a map to not have any duplicates
	testPackages := make(map[string]TestPackage)

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
				//packageName := filepath.Base(dirName)
				packageName, err := getPackageNameFromFile(p)

				if err != nil {
					return err
				}

				testPackages[dirName] = TestPackage{
					Path:        dirName,
					PackageName: packageName,
				}
			}

			return nil
		})
	if err != nil {
		return err
	}

	t1 := template.New("t1")
	t1, err = t1.Parse(`// Code generated by pyroscope.
package {{.PackageName }}

import (
	"github.com/pyroscope-io/client/pyroscope"
)

func init() {
	pyroscope.Start(pyroscope.Config{
		ProfileTypes:    []pyroscope.ProfileType{ {{ .ProfileTypes }} },
		ApplicationName: "{{ .ApplicationName }}",
	})
}
`)
	if err != nil {
		return err
	}

	// For each package
	// Generate a pyroscope_test.go file
	for _, v := range testPackages {
		testFile := path.Join(v.Path, "pyroscope_test.go")
		output, err := os.Create(testFile)
		if err != nil {
			return err
		}
		w := bufio.NewWriter(output)

		err = t1.Execute(w, struct {
			PackageName     string
			ApplicationName string
			ProfileTypes    string
		}{
			PackageName:     v.PackageName,
			ApplicationName: appName,
			ProfileTypes:    strings.Join(profileTypes, ", "),
		})

		if err != nil {
			return err
		}

		if err = w.Flush(); err != nil {
			return err
		}

		fmt.Println("Created file", testFile)
	}

	return nil
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
