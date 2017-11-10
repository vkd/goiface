package main

import (
	"bytes"
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestImpl(t *testing.T) {
	type args struct {
		receiver string
		iface    string
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"base Writer", args{"github.com/vkd/goiface/testdata.MyType", "io.Writer"},
			`func (m MyType) Write(p []byte) (n int, err error) {
	panic("not implemented")
}

`, false},
		{"base ReadWriter", args{"github.com/vkd/goiface/testdata.MyType", "io.ReadWriter"},
			`func (m MyType) Read(p []byte) (n int, err error) {
	panic("not implemented")
}

func (m MyType) Write(p []byte) (n int, err error) {
	panic("not implemented")
}

`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := Impl(tt.args.receiver, tt.args.iface, w); (err != nil) != tt.wantErr {
				t.Errorf("Impl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Impl() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestParsePkgType(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		want  *Pkg
		want1 *Type
	}{
		// TODO: Add test cases.
		{"io.Writer", "io.Writer", &Pkg{Name: "io", Path: "io"}, &Type{Name: "Writer"}},
		{"net/http.Handler", "net/http.Handler", &Pkg{Name: "http", Path: "net/http"}, &Type{Name: "Handler"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParsePkgType(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePkgType() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ParsePkgType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTypeFuncs(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"base type", "github.com/vkd/goiface/testdata.MyType", []string{"MyFunc1", "MyFunc2"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TypeFuncs(ParsePkgType(tt.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("TypeFuncs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var names []string
			for _, fn := range got {
				names = append(names, fn.Name)
			}
			if !reflect.DeepEqual(names, tt.want) {
				t.Errorf("TypeFuncs() = %v, want %v", names, tt.want)
			}
		})
	}
}

func TestIfaceFuncs(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"base interface", "github.com/vkd/goiface/testdata.MyIface", []string{"Iface"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IfaceFuncs(ParsePkgType(tt.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("IfaceFuncs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var names []string
			for _, fn := range got {
				names = append(names, fn.Name)
			}
			if !reflect.DeepEqual(names, tt.want) {
				t.Errorf("IfaceFuncs() = %v, want %v", names, tt.want)
			}
		})
	}
}

func TestParsePkg(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want *Pkg
	}{
		// TODO: Add test cases.
		{"io", "io", &Pkg{Name: "io", Path: "io"}},
		{"http", "net/http", &Pkg{Name: "http", Path: "net/http"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParsePkg(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePkg() = %v, want %v", got, tt.want)
			}
		})
	}
}

type counter struct {
	i int
}

func (c *counter) Each(fset *token.FileSet, d ast.Decl) error {
	c.i++
	return nil
}

func TestPkg_EachDecl(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		count   int
		wantErr bool
	}{
		// TODO: Add test cases.
		{"base", "github.com/vkd/goiface/testdata", 5, false},
		{"base", "github.com/vkd/goiface/notfound", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsePkg(tt.s)
			cnt := counter{}
			if err := p.EachDecl(cnt.Each); (err != nil) != tt.wantErr {
				t.Errorf("Pkg.EachDecl() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cnt.i != tt.count {
				t.Errorf("Wrong decl count: %d (need: %d)", cnt.i, tt.count)
			}
		})
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want *Type
	}{
		// TODO: Add test cases.
		{"base type", "myType", &Type{Name: "myType"}},
		{"pointer type", "*myType", &Type{Name: "myType", IsPtr: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseType(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestType_VarName(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		// TODO: Add test cases.
		{"base type", "myType", "m"},
		{"empty type", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := ParseType(tt.s)
			if got := tp.VarName(); got != tt.want {
				t.Errorf("Type.VarName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestType_VarType(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		// TODO: Add test cases.
		{"base type", "myType", "myType"},
		{"pointer type", "*myType", "*myType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := ParseType(tt.s)
			if got := tp.VarType(); got != tt.want {
				t.Errorf("Type.VarType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunc_Decl(t *testing.T) {

	tests := []struct {
		name    string
		pkg     string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"handler current package", "net/http", "ServeHTTP(ResponseWriter, *Request)", false},
		{"handler another package", "", "ServeHTTP(http.ResponseWriter, *http.Request)", false},
	}

	funcs, err := IfaceFuncs(ParsePkgType("net/http.Handler"))
	if err != nil {
		t.Fatalf("Error on get funcs: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("Count funcs of net/http.Handler not equals 1: %v", funcs)
	}
	handlerFn := funcs[0]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handlerFn.Decl(ParsePkg(tt.pkg))
			if (err != nil) != tt.wantErr {
				t.Errorf("Func.Decl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Func.Decl() = %v, want %v", got, tt.want)
			}
		})
	}
}
