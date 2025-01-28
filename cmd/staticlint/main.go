// Package main содержит точку входа для запуска multichecker.
// Он регистрирует стандартные анализаторы из golang.org/x/tools/go/analysis/passes,
// все анализаторы класса SA из пакета staticcheck.io,
// как минимум один анализатор из остальных классов пакета staticcheck.io,
// два публичных анализатора на выбор,
// а также собственный анализатор, запрещающий os.Exit в main.
package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
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
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	// Анализаторы из staticcheck.io
	"honnef.co/go/tools/staticcheck"

	// Дополнительные публичные анализаторы
	"github.com/timakin/bodyclose/passes/bodyclose"
	"github.com/gostaticanalysis/nilerr"

	// Собственный анализатор
	"github.com/Grifonhard/Practicum-metrics/cmd/staticlint/noexit"
)

func main() {
	// Соберём все анализаторы в один срез
	var analyzers []*analysis.Analyzer

	// 1. Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		inspect.Analyzer,
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
		unusedresult.Analyzer,
	)

	// 2. Анализаторы класса SA из пакета staticcheck.io
	//    (в официальной документации staticcheck они перечислены как SAxxxxx)
	for _, scAnalyzer := range staticcheck.Analyzers {
		if strings.HasPrefix(scAnalyzer.Analyzer.Name, "SA") {
			analyzers = append(analyzers, scAnalyzer.Analyzer)
		}
	}

	// 3. Добавим как минимум один анализатор из остальных классов staticcheck
	//    (например, ST1000 — "Check package comment")
	for _, scAnalyzer := range staticcheck.Analyzers {
		// Выберем стиль-анализатор из группы ST
		if scAnalyzer.Analyzer.Name == "ST1000" {
			analyzers = append(analyzers, scAnalyzer.Analyzer)
			break
		}
	}

	// 4. Добавим два публичных анализатора (пример: gostaticanalysis)
	analyzers = append(analyzers,
		bodyclose.Analyzer, // проверяет комментарии
		nilerr.Analyzer,  // проверяет обработку ошибок nil
	)

	// 5. Добавляем собственный анализатор, запрещающий os.Exit в main
	analyzers = append(analyzers, noexit.MyAnalyzer)

	// Запускаем наш multichecker
	multichecker.Main(analyzers...)
}
