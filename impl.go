package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Impl - create missed funcs in receiver from interface
//
// receiver - example: "github.com/name/project.MyType"
// iface - example: "io.ReadWriter", "github.com/name/project.MyIface"
// w - write a result
func Impl(receiver string, iface string, w io.Writer) error {
	// get funcs by receiver type
	rp, rt := ParsePkgType(receiver)
	rFns, err := TypeFuncs(rp, rt)
	if err != nil {
		return errors.Wrapf(err, "error on get funcs by type (%s)", receiver)
	}

	// make map by receiver's funcs
	mapRecFns := make(map[string]*Func)
	for _, fn := range rFns {
		mapRecFns[fn.Name] = fn
	}

	// get funcs by interface type
	iFns, err := IfaceFuncs(ParsePkgType(iface))
	if err != nil {
		return errors.Wrapf(err, "error on get funcs by iface (%s)", iface)
	}

	// filter interface funcs by existings of receiver
	var newFns []*Func
	for _, fn := range iFns {
		if _, ok := mapRecFns[fn.Name]; !ok {
			newFns = append(newFns, fn)
		}
	}

	// write result by template
	var v = tplVar{
		VarReceiver: rt.VarName(),
		Receiver:    rt.VarType(),
	}
	for i, fn := range newFns {
		name, s, err := fn.Decl(rp)
		if err != nil {
			return errors.Wrapf(err, "error on get decl by func (%s)", fn.Name)
		}
		v.Funcs = append(v.Funcs, tplFunc{Name: name, Decl: s, IsLast: i == len(newFns)-1})
	}
	err = tpl.Execute(w, v)
	if err != nil {
		return errors.Wrap(err, "error on exec template")
	}

	return nil
}

// ParsePkgType - parse string to Pkg and Type
//
// Examples:
// 'io.Writer'
// 'net/http.Handler'
// 'github.com/user/project.MyIface'
// 'myLocalIface'
func ParsePkgType(s string) (*Pkg, *Type) {
	idx := strings.LastIndex(s, ".")
	if idx < 0 {
		return ParsePkg(""), ParseType(s)
	}
	return ParsePkg(s[:idx]), ParseType(s[idx+1:])
}

// TypeFuncs - return funcs by type
func TypeFuncs(p *Pkg, t *Type) ([]*Func, error) {
	var out []*Func

	err := p.EachDecl(func(fset *token.FileSet, d ast.Decl) error {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			return nil
		}
		if fd.Recv == nil {
			return nil
		}
		switch tp := fd.Recv.List[0].Type.(type) {
		case *ast.Ident: // func (myType) Name() {...}
			if tp.Name == t.Name {
				out = append(out, &Func{Name: fd.Name.Name})
			}
		case *ast.StarExpr: // func (*myType) Name() {...}
			if ident, ok := tp.X.(*ast.Ident); ok {
				if ident.Name == t.Name {
					out = append(out, &Func{Name: fd.Name.Name})
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "error on each pkg")
	}

	return out, nil
}

// IfaceFuncs - return funcs by interface
func IfaceFuncs(p *Pkg, t *Type) ([]*Func, error) {
	var out []*Func

	err := p.EachDecl(func(fset *token.FileSet, d ast.Decl) error {
		gd, ok := d.(*ast.GenDecl)
		if !ok {
			return nil
		}
		for _, sp := range gd.Specs {
			sp, ok := sp.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if sp.Name.Name != t.Name {
				continue
			}
			t, ok := sp.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, m := range t.Methods.List {
				if len(m.Names) == 0 { // embedded
					id, ok := m.Type.(*ast.Ident)
					if !ok {
						continue
					}
					ep, et := ParsePkgType(id.Name)
					if ep.Path == "" { // if embedded type in current package
						ep = p
					}
					fns, err := IfaceFuncs(ep, et)
					if err != nil {
						return errors.Wrapf(err, "error on embedded interface (%s)", id.Name)
					}
					out = append(out, fns...)
					continue
				}
				ftype, ok := m.Type.(*ast.FuncType)
				if !ok {
					continue
				}

				out = append(out, &Func{
					Name:  m.Names[0].Name,
					Pkg:   p,
					fset:  fset,
					ftype: ftype,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "error on each pkg")
	}

	return out, nil
}

// Pkg - golang package
type Pkg struct {
	Name string // 'http'
	Path string // 'net/http'
}

// ParsePkg - parse golang package
func ParsePkg(s string) *Pkg {
	idx := strings.LastIndex(s, "/")
	if idx < 0 {
		return &Pkg{Name: s, Path: s}
	}
	return &Pkg{Name: s[idx+1:], Path: s}
}

// EachFn - func type for walk by ast tree
type EachFn func(fset *token.FileSet, d ast.Decl) error

// EachDecl - walk by package declarations
func (p *Pkg) EachDecl(fn EachFn) error {
	pkg, err := build.Import(p.Path, "", 0)
	if err != nil {
		return errors.Wrapf(err, "error on import go pkg (%s)", p.Path)
	}

	fset := token.NewFileSet()
	for _, goFile := range pkg.GoFiles {
		goFile = filepath.Join(pkg.Dir, goFile)
		f, err := parser.ParseFile(fset, goFile, nil, 0)
		if err != nil {
			return errors.Wrapf(err, "error on parse go file (%s)", goFile)
		}
		for _, d := range f.Decls {
			err = fn(fset, d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Type - golang type
type Type struct {
	Name  string
	IsPtr bool
}

// ParseType - parse golang type
func ParseType(s string) *Type {
	isPtr := strings.HasPrefix(s, "*")
	if isPtr {
		s = s[1:]
	}
	return &Type{s, isPtr}
}

// VarName - return recomended name of golang type
//
// using first sym from typename and convert to lower
func (t *Type) VarName() string {
	if len(t.Name) > 0 {
		return strings.ToLower(string(t.Name[0]))
	}
	return ""
}

// VarType - return "*<name>" if is ptr
func (t *Type) VarType() string {
	name := t.Name
	if t.IsPtr {
		name = "*" + name
	}
	return name
}

// Func - golang func
type Func struct {
	Name string
	Pkg  *Pkg

	fset  *token.FileSet
	ftype *ast.FuncType
}

// Decl - string representation of func on target package
func (f *Func) Decl(targetPkg *Pkg) (string, string, error) {
	// if func from another package
	if targetPkg.Path != f.Pkg.Path {
		ast.Inspect(f.ftype, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				if n.IsExported() {
					n.Name = f.Pkg.Name + "." + n.Name
				}
			}
			return true
		})
	}

	funcVar := &ast.FuncDecl{
		Name: &ast.Ident{Name: f.Name},
		Type: f.ftype,
	}

	var bs bytes.Buffer
	// format to pretty string
	err := format.Node(&bs, f.fset, funcVar)
	if err != nil {
		return "", "", errors.Wrap(err, "error on format func")
	}
	return f.Name, strings.TrimPrefix(bs.String(), "func "), nil
}

var (
	tpl = mustTemplate(`
{{- $varR := .VarReceiver}}
{{- $rec := .Receiver}}
{{- range .Funcs -}}
// {{.Name}} ...
func ({{$varR}} {{$rec}}) {{.Decl}} {
	panic("not implemented")
}
{{if not .IsLast}}
{{end -}}
{{end -}}
`)
)

func mustTemplate(s string) *template.Template {
	return template.Must(template.New("").Parse(s))
}

type tplVar struct {
	VarReceiver string
	Receiver    string
	Funcs       []tplFunc
}

type tplFunc struct {
	Name   string
	Decl   string
	IsLast bool
}
