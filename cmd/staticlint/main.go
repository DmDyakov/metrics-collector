// Package main реализует утилиту статического анализа кода, объединяющую
// стандартные анализаторы go vet и дополнительные проверки.
//
// Утилита запускает multichecker со следующим набором анализаторов:
//
// 1. Стандартные анализаторы пакета golang.org/x/tools/go/analysis/passes:
//   - printf - проверка форматных строк
//   - shadow - проверка затенения переменных
//   - assign - проверка присваиваний
//   - и другие анализаторы из пакета go/analysis
//
// 2. Анализаторы класса SA пакета staticcheck.io (все проверки):
//   - Содержит около 80 анализаторов для поиска серьёзных ошибок и багов,
//     включая проверки конкурентности, утечек, неэффективного кода.
//
// 3. Анализаторы других классов staticcheck.io:
//   - ST (style) - стилистические замечания
//   - QF (quickfix) - предложения по быстрому исправлению кода
//
// 4. Публичные анализаторы:
//   - nilerr (github.com/gostaticanalysis/nilerr) - поиск ошибочного
//     возврата nil вместо ошибки (return nil, err)
//   - tenv (github.com/sivchari/tenv) - проверка использования
//     os.Setenv в тестах, требует замены на t.Setenv
//
// 5. Кастомный анализатор:
//   - noosexit - запрещает прямой вызов os.Exit в функции main пакета main.
//     Игнорирует сгенерированные файлы и кэш сборки Go.
//
// Механизм запуска multichecker:
//
// Локальный запуск без сборки:
//
//	go run ./cmd/staticlint ./...
//
// Локальный запуск конкретного пакета:
//
//	go run ./cmd/staticlint ./internal/...
//
// Сборка бинарного файла:
//
//	go build -o staticlint ./cmd/staticlint
//
// Запуск собранного бинарника:
//
//	./staticlint ./...
//
// Запуск с автоматическим исправлением (для QF-анализаторов):
//
//	./staticlint -fix ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/gostaticanalysis/nilerr"
	"github.com/sivchari/tenv"
)

func main() {
	analyzers := []*analysis.Analyzer{
		noOsExitAnalyzer,

		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,

		nilerr.Analyzer,
		tenv.Analyzer,
	}

	for _, sa := range staticcheck.Analyzers {
		analyzers = append(analyzers, sa.Analyzer)
	}
	for _, st := range stylecheck.Analyzers {
		analyzers = append(analyzers, st.Analyzer)
	}
	for _, qf := range quickfix.Analyzers {
		analyzers = append(analyzers, qf.Analyzer)
	}

	multichecker.Main(analyzers...)
}
