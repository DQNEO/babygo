package builder

import (
	"github.com/DQNEO/babygo/internal/compiler"
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
func sortTopologically(tree DependencyTree, prepend []string) []string {
	sorted := prepend
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

	var r []string
	var users []string

	for _, pth := range sorted {
		if isStdLib(pth) {
			r = append(r, pth)
		} else {
			users = append(users, pth)
		}
	}
	for _, pth := range users {
		r = append(r, pth)
	}
	return r
}

func (b *Builder) getPackageDir(importPath string) string {
	if strings.HasPrefix(importPath, "./") {
		// relative path means it is the package directory
		return importPath
	}
	if isStdLib(importPath) {
		return b.BbgRootSrcPath + "/" + importPath
	} else {
		return b.SrcPath + "/" + importPath
	}
}

func (b *Builder) getPackageSourceFiles(pkgPath string) []string {
	packageDir := b.getPackageDir(pkgPath)
	files := findFilesInDir(packageDir, true)
	if len(files) == 0 {
		panic("No source files found in " + packageDir + " (" + pkgPath + ")")
	}
	return files
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

func (b *Builder) ListDepth(workdir string, pkgPath string) {
	b.filesCache = make(map[string][]string)
	b.permanentTree = make(map[string]*compiler.PackageToCompile)

	files := b.getPackageSourceFiles(pkgPath)
	var gofiles []string
	var asmfiles []string
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			gofiles = append(gofiles, file)
		} else if strings.HasSuffix(file, ".s") {
			asmfiles = append(asmfiles, file)
		}
	}

	var mainGoFiles []string = gofiles

	imports := collectImportsFromFiles(mainGoFiles)
	importsList := mapToSlice(imports)
	b.permanentTree["main"] = &compiler.PackageToCompile{
		Name:    "main",
		Path:    "main",
		GoFiles: mainGoFiles,
		Imports: importsList,
	}
	imports["runtime"] = true
	tree := make(DependencyTree)
	b.collectDependency(tree, imports)
	prepend := []string{"unsafe", "runtime"}
	sortedPaths := sortTopologically(tree, prepend)
	sortedPaths = append(sortedPaths, "main")

	for _, pth := range sortedPaths {
		pkg := b.permanentTree[pth]
		fmt.Printf("%s\n", pkg.Path)
	}
}

func (b *Builder) BuildOne(workdir string, outputBaseName string, pkgPath string) {
	b.filesCache = make(map[string][]string)
	b.permanentTree = make(map[string]*compiler.PackageToCompile)

	files := b.getPackageSourceFiles(pkgPath)
	var gofiles []string
	var asmfiles []string
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			gofiles = append(gofiles, file)
		} else if strings.HasSuffix(file, ".s") {
			asmfiles = append(asmfiles, file)
		}
	}

	var pkgName string
	if strings.HasPrefix(pkgPath, "./") {
		pkgName = "main"
	} else {
		pkgName = path.Base(pkgPath)
	}

	imports := collectImportsFromFiles(gofiles)
	importsList := mapToSlice(imports)
	pkg := &compiler.PackageToCompile{
		Name:     pkgName,
		Path:     pkgPath,
		GoFiles:  gofiles,
		AsmFiles: asmfiles,
		Imports:  importsList,
	}

	var prepend []string

	switch pkgPath {
	case "unsafe":
		// do not prepend
	case "runtime":
		prepend = []string{"unsafe"}
	default:
		prepend = []string{"unsafe", "runtime"}
		imports["runtime"] = true
	}

	tree := make(DependencyTree)
	b.collectDependency(tree, imports)
	sortedPaths := sortTopologically(tree, prepend)

	var uni = universe.CreateUniverse()
	sema.Fset = token.NewFileSet()

	for _, path := range sortedPaths {
		basename := normalizeImportPath(path)
		declFilePath := fmt.Sprintf("%s/%s", workdir, basename+".dcl.go")
		compiler.CompileDecl(uni, sema.Fset, path, declFilePath)
	}
	outAsmPath := outputBaseName + ".s"
	declFilePath := outputBaseName + ".dcl.go"
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
