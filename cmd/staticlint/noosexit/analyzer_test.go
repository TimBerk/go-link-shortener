package noosexit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		pkgName string
		code    string
		want    bool
		wantMsg string
	}{
		{
			name:    "os.Exit in main package",
			pkgName: "main",
			code:    `os.Exit(1)`,
			want:    true,
			wantMsg: "direct call to os.Exit in main package is forbidden",
		},
		{
			name:    "os.Exit in non-main package",
			pkgName: "other",
			code:    `os.Exit(1)`,
			want:    false,
		},
		{
			name:    "other function in main package",
			pkgName: "main",
			code:    `fmt.Println("hello")`,
			want:    false,
		},
		{
			name:    "Exit from different package",
			pkgName: "main",
			code:    `otherpkg.Exit(1)`,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSet := token.NewFileSet()
			f, err := parser.ParseFile(fileSet, "test.go", "package "+tt.pkgName+"\nfunc main() {"+tt.code+"}", 0)
			if err != nil {
				t.Fatalf("Failed to parse test code: %v", err)
			}

			testInspector := inspector.New([]*ast.File{f})

			// Create a minimal types.Package implementation
			pkg := types.NewPackage("", tt.pkgName)

			// Prepare analysis pass
			var got bool
			var gotMsg string
			pass := &analysis.Pass{
				Analyzer: Analyzer,
				ResultOf: map[*analysis.Analyzer]interface{}{
					inspect.Analyzer: testInspector,
				},
				Report: func(d analysis.Diagnostic) {
					got = true
					gotMsg = d.Message
				},
				Fset:      fileSet,
				Files:     []*ast.File{f},
				Pkg:       pkg,
				TypesInfo: &types.Info{},
			}

			_, err = run(pass)
			if err != nil {
				t.Fatalf("run() failed: %v", err)
			}

			if got != tt.want {
				t.Errorf("run() report = %v, want %v", got, tt.want)
			}
			if got && gotMsg != tt.wantMsg {
				t.Errorf("run() message = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}
