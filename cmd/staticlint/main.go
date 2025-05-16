// Package main отвечает за конфигурирование, запуск и работу нескольких анализаторов
package main

import (
	"go/ast"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"honnef.co/go/tools/analysis/lint"

	"github.com/TimBerk/go-link-shortener/cmd/staticlint/noosexit"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// isBenchmarkFile проверяет является ли файл Benchmark-тестом
func isBenchmarkFile(pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		fileName := filepath.Base(pass.Fset.File(f.Pos()).Name())
		if strings.HasSuffix(fileName, "_benchmark_test.go") {
			return true
		}
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "Benchmark") {
				return true
			}
		}
	}
	return false
}

// wrapRun создает обернутую функцию Run
func wrapRun(analyzer *analysis.Analyzer) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		if isBenchmarkFile(pass) {
			if analyzer.ResultType != nil {
				return reflect.New(analyzer.ResultType).Elem().Interface(), nil
			}
			return nil, nil
		}
		return analyzer.Run(pass)
	}
}

func findAnalyzer(analyzers []*lint.Analyzer, name string) *analysis.Analyzer {
	for _, a := range analyzers {
		if a.Analyzer.Name == name {
			return a.Analyzer
		}
	}
	return nil
}

func main() {
	var analyzers []*analysis.Analyzer

	// Добавляем стандартные анализаторы
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	// Добавляем анализаторы staticcheck
	for _, v := range staticcheck.Analyzers {
		// Добавляем все анализаторы класса SA
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Добавляем по одному анализатору из других классов
	if a := findAnalyzer(simple.Analyzers, "S1000"); a != nil {
		analyzers = append(analyzers, a)
	}

	if a := findAnalyzer(stylecheck.Analyzers, "ST1000"); a != nil {
		analyzers = append(analyzers, a)
	}

	// Добавляем дополнительные анализаторы
	errCheckAnalyzer := errcheck.Analyzer
	analyzers = append(analyzers, &analysis.Analyzer{
		Name:       errCheckAnalyzer.Name,
		Doc:        errCheckAnalyzer.Doc,
		Run:        wrapRun(errCheckAnalyzer),
		Requires:   errCheckAnalyzer.Requires,
		FactTypes:  errCheckAnalyzer.FactTypes,
		ResultType: errCheckAnalyzer.ResultType,
	})

	// Добавляем собственный анализатор
	analyzers = append(analyzers, noosexit.Analyzer)

	multichecker.Main(analyzers...)
}
