package builder

import (
	"os"

	"github.com/DQNEO/babygo/internal/codegen"
	"github.com/DQNEO/babygo/internal/compiler"
	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/parser"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strings"
	"github.com/DQNEO/babygo/lib/token"
)

type PackageToBuild struct {
	path  string
	name  string
	files []string
}

// "some/dir" => []string{"a.go", "b.go"}
func findFilesInDir(dir string) []string {
	dirents, _ := mylib.Readdirnames(dir)
	var r []string
	for _, dirent := range dirents {
		if dirent == "_.s" {
			continue
		}
		if strings.HasSuffix(dirent, ".go") || strings.HasSuffix(dirent, ".s") {
			r = append(r, dirent)
		}
	}
	return r
}

func isStdLib(pth string) bool {
	return !strings.Contains(pth, ".")
}

func getImportPathsFromFile(file string) []string {
	fset := &token.FileSet{}
	astFile0 := parseImports(fset, file)
	var paths []string
	for _, importSpec := range astFile0.Imports {
		rawValue := importSpec.Path.Value
		pth := rawValue[1 : len(rawValue)-1]
		paths = append(paths, pth)
	}
	return paths
}

type DependencyTree map[string]map[string]bool

func removeNode(tree DependencyTree, node string) {
	for _, paths := range tree {
		delete(paths, node)
	}

	delete(tree, node)
}

func getKeys(tree DependencyTree) []string {
	var keys []string
	for k, _ := range tree {
		keys = append(keys, k)
	}
	return keys
}

// Do topological sort
// In the result list, the independent (lowest level) packages come first.
func sortTopologically(tree DependencyTree) []string {
	var sorted []string
	for len(tree) > 0 {
		keys := getKeys(tree)
		mylib.SortStrings(keys)
		for _, _path := range keys {
			children, ok := tree[_path]
			if !ok {
				panic("not found in tree")
			}
			if len(children) == 0 {
				// collect leaf node
				sorted = append(sorted, _path)
				removeNode(tree, _path)
			}
		}
	}
	return sorted
}

func getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return BbgRootSrcPath + "/" + importPath
	} else {
		return SrcPath + "/" + importPath
	}
}

func collectDependency(tree DependencyTree, paths map[string]bool) {
	for pkgPath, _ := range paths {
		if pkgPath == "unsafe" || pkgPath == "runtime" {
			continue
		}
		packageDir := getPackageDir(pkgPath)
		fnames := findFilesInDir(packageDir)
		children := make(map[string]bool)
		for _, fname := range fnames {
			if !strings.HasSuffix(fname, ".go") {
				// skip ".s"
				continue
			}
			_paths := getImportPathsFromFile(packageDir + "/" + fname)
			for _, pth := range _paths {
				if pth == "unsafe" || pth == "runtime" {
					continue
				}
				children[pth] = true
			}
		}
		tree[pkgPath] = children
		collectDependency(tree, children)
	}
}

func collectAllPackages(inputFiles []string) []string {
	directChildren := collectDirectDependents(inputFiles)
	tree := make(DependencyTree)
	collectDependency(tree, directChildren)
	sortedPaths := sortTopologically(tree)

	// sort packages by this order
	// 1: pseudo
	// 2: stdlib
	// 3: external
	paths := []string{"unsafe", "runtime"}
	for _, pth := range sortedPaths {
		if isStdLib(pth) {
			paths = append(paths, pth)
		}
	}
	for _, pth := range sortedPaths {
		if !isStdLib(pth) {
			paths = append(paths, pth)
		}
	}
	return paths
}

func collectDirectDependents(inputFiles []string) map[string]bool {
	importPaths := make(map[string]bool)
	for _, inputFile := range inputFiles {
		paths := getImportPathsFromFile(inputFile)
		for _, pth := range paths {
			importPaths[pth] = true
		}
	}
	return importPaths
}

func collectSourceFiles(pkgDir string) []string {
	fnames := findFilesInDir(pkgDir)
	var files []string
	for _, fname := range fnames {
		srcFile := pkgDir + "/" + fname
		files = append(files, srcFile)
	}
	return files
}

func parseImports(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		panic(filename + ":" + err.Error())
	}
	return f
}

var SrcPath string
var BbgRootSrcPath string

func BuildAll(srcPath string, bbgRootSrcPath string, workdir string, args []string) {
	SrcPath = srcPath
	BbgRootSrcPath = bbgRootSrcPath

	var inputFiles []string
	for _, arg := range args {
		switch arg {
		case "-DG":
			codegen.DebugCodeGen = true
		default:
			inputFiles = append(inputFiles, arg)
		}
	}

	paths := collectAllPackages(inputFiles)
	var packagesToBuild []*PackageToBuild
	for _, _path := range paths {
		files := collectSourceFiles(getPackageDir(_path))
		packagesToBuild = append(packagesToBuild, &PackageToBuild{
			name:  path.Base(_path),
			path:  _path,
			files: files,
		})
	}

	packagesToBuild = append(packagesToBuild, &PackageToBuild{
		name:  "main",
		path:  "main",
		files: inputFiles,
	})

	var universe *ast.Scope = universe.CreateUniverse()
	sema.Fset = token.NewFileSet()
	var builtPackages []*ir.AnalyzedPackage
	for _, _pkg := range packagesToBuild {
		if _pkg.name == "" {
			panic("empty pkg name")
		}
		var asmBasename []byte
		for _, ch := range []byte(_pkg.path) {
			if ch == '/' {
				ch = '.'
			}
			asmBasename = append(asmBasename, ch)
		}
		outFilePath := fmt.Sprintf("%s/%s", workdir, string(asmBasename)+".s")
		var gofiles []string
		var asmfiles []string
		for _, f := range _pkg.files {
			if strings.HasSuffix(f, ".go") {
				gofiles = append(gofiles, f)
			} else if strings.HasSuffix(f, ".s") {
				asmfiles = append(asmfiles, f)
			}

		}
		apkg := compiler.Compile(universe, sema.Fset, _pkg.path, _pkg.name, gofiles, asmfiles, outFilePath)
		builtPackages = append(builtPackages, apkg)
	}

	outFilePath := fmt.Sprintf("%s/%s", workdir, "__INIT__.s")
	outAsmFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(outAsmFile, ".text\n")
	fmt.Fprintf(outAsmFile, "# Initializes all packages except for runtime\n")
	fmt.Fprintf(outAsmFile, ".global __INIT__.init\n")
	fmt.Fprintf(outAsmFile, "__INIT__.init:\n")
	for _, _pkg := range builtPackages {
		// A package with no imports is initialized by assigning initial values to all its package-level variables
		//  followed by calling all init functions in the order they appear in the source

		if _pkg.Name == "runtime" {
			// Do not call runtime inits. They should be called manually from runtime asm
			continue
		}

		fmt.Fprintf(outAsmFile, "  callq %s.__initVars \n", _pkg.Name)
		if _pkg.HasInitFunc {
			fmt.Fprintf(outAsmFile, "  callq %s.init \n", _pkg.Name)
		}
	}
	fmt.Fprintf(outAsmFile, "  ret\n")
	outAsmFile.Close()
}
