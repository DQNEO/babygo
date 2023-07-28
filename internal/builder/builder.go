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
	var sorted []string = []string{"unsafe", "runtime"}
	removeNode(tree, "unsafe")
	removeNode(tree, "runtime")
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
		_, ok := b.filesCache[pkgPath]
		if ok {
			continue
		}
		files := b.getPackageSourceFiles(pkgPath)
		b.filesCache[pkgPath] = files
		var gofiles []string
		var asmfiles []string
		for _, file := range files {
			if strings.HasSuffix(file, ".go") {
				gofiles = append(gofiles, file)
			} else if strings.HasSuffix(file, ".s") {
				asmfiles = append(asmfiles, file)
			}
		}

		imports := collectImportsFromFiles(gofiles)
		importsList := mapToSlice(imports)

		tree[pkgPath] = imports
		b.permanentTree[pkgPath] = &compiler.PackageToCompile{
			Path:     pkgPath,
			Name:     path.Base(pkgPath),
			GoFiles:  gofiles,
			AsmFiles: asmfiles,
			Imports:  importsList,
		}

		b.collectDependency(tree, imports)
	}
}

func mapToSlice(imports map[string]bool) []string {
	var list []string
	for k, _ := range imports {
		list = append(list, k)
	}
	return list
}

func collectImportsFromFiles(gofiles []string) map[string]bool {
	imports := make(map[string]bool)
	for _, gofile := range gofiles {
		pths := getImportPathsFromFile(gofile)
		for _, pth := range pths {
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
	filesCache     map[string][]string
	permanentTree  map[string]*compiler.PackageToCompile
}

func (b *Builder) Build(workdir string, args []string) {
	b.filesCache = make(map[string][]string)
	b.permanentTree = make(map[string]*compiler.PackageToCompile)

	var mainFiles []string
	for _, arg := range args {
		switch arg {
		case "-DG":
			codegen.DebugCodeGen = true
		default:
			mainFiles = append(mainFiles, arg)
		}
	}

	imports := collectImportsFromFiles(mainFiles)
	importsList := mapToSlice(imports)
	b.permanentTree["main"] = &compiler.PackageToCompile{
		Name:    "main",
		Path:    "main",
		GoFiles: mainFiles,
		Imports: importsList,
	}
	imports["runtime"] = true
	tree := make(DependencyTree)
	b.collectDependency(tree, imports)
	sortedPaths := sortTopologically(tree)
	sortedPaths = append(sortedPaths, "main")

	var uni = universe.CreateUniverse()
	sema.Fset = token.NewFileSet()

	var builtPackages []*ir.AnalyzedPackage
	for _, path := range sortedPaths {
		pkg := b.permanentTree[path]
		fmt.Fprintf(os.Stderr, "Building  %s %s\n", pkg.Name, pkg.Path)
		basename := normalizeImportPath(pkg.Path)
		outAsmPath := fmt.Sprintf("%s/%s", workdir, basename+".s")
		declFilePath := fmt.Sprintf("%s/%s", workdir, basename+".dcl.go")
		apkg := compiler.Compile(uni, sema.Fset, pkg, outAsmPath, declFilePath)
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

func (b *Builder) Compile(workdir string, args []string) {
	_ = args[0] // -o
	outputBaseName := args[1]
	pkgPath := args[2]

	b.filesCache = make(map[string][]string)
	b.permanentTree = make(map[string]*compiler.PackageToCompile)

	fmt.Fprintf(os.Stderr, "Compiling  %s ...\n", pkgPath)
	files := b.getPackageSourceFiles(pkgPath)
	for _, file := range files {
		fmt.Fprintf(os.Stderr, "  %s\n", file)
	}
	var gofiles []string
	var asmfiles []string
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			gofiles = append(gofiles, file)
		} else if strings.HasSuffix(file, ".s") {
			asmfiles = append(asmfiles, file)
		}
	}

	imports := collectImportsFromFiles(gofiles)
	importsList := mapToSlice(imports)
	pkg := &compiler.PackageToCompile{
		Name:    path.Base(pkgPath),
		Path:    pkgPath,
		GoFiles: gofiles,
		Imports: importsList,
	}
	imports["runtime"] = true
	tree := make(DependencyTree)
	b.collectDependency(tree, imports)
	sortedPaths := sortTopologically(tree)
	sortedPaths = append(sortedPaths, pkgPath)

	var uni = universe.CreateUniverse()
	sema.Fset = token.NewFileSet()

	for _, path := range sortedPaths {
		fmt.Fprintf(os.Stderr, "  import %s\n", path)
		//@TODO:  compiler.CompileDec()
		basename := normalizeImportPath(path)
		declFilePath := fmt.Sprintf("%s/%s", workdir, basename+".dcl.go")
		compiler.CompileDecl(uni, sema.Fset, path, declFilePath)
	}
	fmt.Fprintf(os.Stderr, "outputFile %s\n", outputBaseName)
	outAsmPath := outputBaseName + ".2s"
	declFilePath := outputBaseName + ".2dcl.go"
	compiler.Compile(uni, sema.Fset, pkg, outAsmPath, declFilePath)
}

func normalizeImportPath(importPath string) string {
	return strReplace(importPath, '/', '.')
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
