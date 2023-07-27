package builder

import (
	"os"

	"github.com/DQNEO/babygo/internal/codegen"
	"github.com/DQNEO/babygo/internal/compiler"
	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/internal/types"
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

func init() {
	// Check object addresses
	tIdent := types.Int.E.(*ast.Ident)
	if tIdent.Obj != universe.Int {
		panic("object mismatch")
	}

}

// "some/dir" => "some/dir/a.go" (if abs == true)
func findFilesInDir(dir string, abs bool) []string {
	dirents, _ := mylib.Readdirnames(dir)
	var r []string
	for _, dirent := range dirents {
		if strings.HasSuffix(dirent, ".go") || strings.HasSuffix(dirent, ".s") {
			var file string
			if abs {
				file = dir + "/" + dirent
			} else {
				file = dirent
			}
			r = append(r, file)
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

func (b *Builder) getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return b.BbgRootSrcPath + "/" + importPath
	} else {
		return b.SrcPath + "/" + importPath
	}
}

func (b *Builder) getPackageSourceFiles(pkgPath string) []string {
	packageDir := b.getPackageDir(pkgPath)
	return findFilesInDir(packageDir, true)
}

func (b *Builder) collectDependency(tree DependencyTree, paths map[string]bool) {
	for pkgPath, _ := range paths {
		if pkgPath == "unsafe" || pkgPath == "runtime" {
			continue
		}

		files := b.getPackageSourceFiles(pkgPath)
		var gofiles []string
		for _, file := range files {
			if strings.HasSuffix(file, ".go") {
				gofiles = append(gofiles, file)
			}
		}

		imports := collectImportsFromFiles(gofiles)
		tree[pkgPath] = imports
		b.collectDependency(tree, imports)
	}
}

func (b *Builder) collectAllPackages(mainFiles []string) []string {
	imports := collectImportsFromFiles(mainFiles)
	tree := make(DependencyTree)
	b.collectDependency(tree, imports)
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

func collectImportsFromFiles(gofiles []string) map[string]bool {
	imports := make(map[string]bool)
	for _, gofile := range gofiles {
		pths := getImportPathsFromFile(gofile)
		for _, pth := range pths {
			if pth == "unsafe" || pth == "runtime" {
				continue
			}
			imports[pth] = true
		}
	}
	return imports
}

func parseImports(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		panic(filename + ":" + err.Error())
	}
	return f
}

type Builder struct {
	SrcPath        string // user-land packages
	BbgRootSrcPath string // std packages
}

var Workdir string

func (b *Builder) Build(workdir string, args []string) {
	Workdir = workdir
	var inputFiles []string
	for _, arg := range args {
		switch arg {
		case "-DG":
			codegen.DebugCodeGen = true
		default:
			inputFiles = append(inputFiles, arg)
		}
	}

	var uni = universe.CreateUniverse()
	sema.Fset = token.NewFileSet()

	paths := b.collectAllPackages(inputFiles)
	var packagesToBuild []*PackageToBuild
	for _, _path := range paths {
		files := findFilesInDir(b.getPackageDir(_path), true)
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

	var builtPackages []*ir.AnalyzedPackage
	for _, _pkg := range packagesToBuild {
		if _pkg.name == "" {
			panic("empty pkg name")
		}
		basename := strReplace(_pkg.path, '/', '.')
		outFilePath := fmt.Sprintf("%s/%s", workdir, basename+".s")
		var gofiles []string
		var asmfiles []string
		for _, f := range _pkg.files {
			if strings.HasSuffix(f, ".go") {
				gofiles = append(gofiles, f)
			} else if strings.HasSuffix(f, ".s") {
				asmfiles = append(asmfiles, f)
			}

		}
		apkg := compiler.Compile(uni, sema.Fset, _pkg.path, _pkg.name, gofiles, asmfiles, outFilePath)
		builtPackages = append(builtPackages, apkg)
	}

	// Write to init asm
	outFilePath := fmt.Sprintf("%s/%s", workdir, "__INIT__.s")
	initAsm, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(initAsm, ".text\n")
	fmt.Fprintf(initAsm, "# Initializes all packages except for runtime\n")
	fmt.Fprintf(initAsm, ".global __INIT__.init\n")
	fmt.Fprintf(initAsm, "__INIT__.init:\n")
	for _, _pkg := range builtPackages {
		// A package with no imports is initialized by assigning initial values to all its package-level variables
		//  followed by calling all init functions in the order they appear in the source

		if _pkg.Name == "runtime" {
			// Do not call runtime inits. They should be called manually from runtime asm
			continue
		}

		fmt.Fprintf(initAsm, "  callq %s.__initVars \n", _pkg.Name)
		if _pkg.HasInitFunc { // @TODO: eliminate this flag. We should always call this
			fmt.Fprintf(initAsm, "  callq %s.init \n", _pkg.Name)
		}
	}
	fmt.Fprintf(initAsm, "  ret\n")
	initAsm.Close()
}

// replace a by b in s
func strReplace(s string, a byte, b byte) string {
	var r []byte
	for _, ch := range []byte(s) {
		if ch == '/' {
			ch = '.'
		}
		r = append(r, ch)
	}
	return string(r)
}
