package deb

import (
	"reflect"
	"testing"
)

var pkgTest = []struct {
	path   string
	name   string
	size   int64
	files  []string
	fields map[string]string
}{
	{
		"testdata/basic.deb",
		"basic.deb",
		576,
		[]string{"./a", "./b", "./c"},
		map[string]string{
			"Package":    "empty",
			"Version":    "1.0.0",
			"Maintainer": "Philip Nelson <me@pnelson.ca>",
			"MD5sum":     "bd896c5308d241ec33837f26570e6c3b",
			"SHA1":       "61368e56b850266de1b1556c41edbcefe132dbbd",
			"SHA256":     "2d8ad6e3bf60551b18a9cb2d50bc919809892cf495e54c9b87258a81830c7ef1",
		},
	},
}

func TestNewPackage(t *testing.T) {
	for i, tt := range pkgTest {
		pkg, err := NewPackage(tt.path)
		if pkg == nil {
			t.Fatalf("%d. problem opening package %v", i, err)
		}

		if pkg.Path != tt.path {
			t.Errorf("%d. pkg.Path have %q, want %q", i, pkg.Path, tt.path)
		}

		if pkg.Name != tt.name {
			t.Errorf("%d. pkg.Name have %q, want %q", i, pkg.Name, tt.name)
		}

		if pkg.Size != tt.size {
			t.Errorf("%d. pkg.Size have %d, want %d", i, pkg.Size, tt.size)
		}

		if !reflect.DeepEqual(pkg.Files, tt.files) {
			t.Errorf("%d. pkg.Files\nhave: %+v\nwant: %+v", i, pkg.Files, tt.files)
		}

		if !reflect.DeepEqual(pkg.Fields, tt.fields) {
			t.Errorf("%d. pkg.Fields\nhave: %+v\nwant: %+v", i, pkg.Fields, tt.fields)
		}
	}
}
