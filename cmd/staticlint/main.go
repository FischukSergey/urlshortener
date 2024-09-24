// Multichecker состоящий из стандартных статических анализаторов пакета
// golang.org/x/tools/go/analysis/passes и анализаторов из пакета staticcheck.io.
// Полный список анализаторов из staticcheck.io можно найти в файле staticcheck.yaml
// Кроме того, добавлены анализаторы, обнаруживающие вызовы os.Exit в функции main.
//
// Использование:
//
// В корне проекта запустите 'make my-lint'.
// Команда создаст файл result.txt с результатами анализа.
// Для удаления файла result.txt запустить комманду 'make clear-my-lint'
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"

	customAnalysis "github.com/FischukSergey/urlshortener.git/cmd/staticlint/mycheck"
)

func main() {
	staticChecks := []*analysis.Analyzer{
		// обнаруживает, если в append есть только один аргумент
		appends.Analyzer,
		// сообщает о несоответствиях между файлами сборки и объявлениями Go.
		asmdecl.Analyzer,
		// обнаруживает бесполезные присваивания.
		assign.Analyzer,
		// проверяет распространенные ошибки при использовании пакета sync/atomic.
		atomic.Analyzer,
		// проверяет аргументы функций sync/atomic на выравнивание по 64-битной границе.
		atomicalign.Analyzer,
		// обнаруживает распространенные ошибки, связанные с булевыми операторами.
		bools.Analyzer,
		// строит представление SSA безошибочного пакета и возвращает набор всех функций в нем.
		buildssa.Analyzer,
		// проверяет теги сборки.
		buildtag.Analyzer,
		// обнаруживает некоторые нарушения правил передачи указателей cgo.
		cgocall.Analyzer,
		// проверяет литералы составных типов без ключей.
		composite.Analyzer,
		// проверяет ошибочную передачу блокировок по значению.
		copylock.Analyzer,
		// предоставляет синтаксический граф управления потоком (CFG) для тела функции.
		ctrlflow.Analyzer,
		// проверяет использование reflect.DeepEqual с значениями ошибок.
		deepequalerrors.Analyzer,
		// проверяет распространенные ошибки в инструкциях defer.
		defers.Analyzer,
		// проверяет известные директивы инструментария Go.
		directive.Analyzer,
		// проверяет, что второй аргумент errors.As является указателем на тип, реализующий error.
		errorsas.Analyzer,
		// обнаруживает структуры, которые использовали бы меньше памяти, если бы их поля были отсортированы.
		fieldalignment.Analyzer,
		// служит тривиальным примером и тестом API анализа.
		findcall.Analyzer,
		// сообщает о коде сборки, который уничтожает указатель на фрейм до его сохранения.
		framepointer.Analyzer,
		// проверяет ошибки при использовании HTTP-ответов.
		httpresponse.Analyzer,
		// выделяет невозможные утверждения типа интерфейс-интерфейс.
		ifaceassert.Analyzer,
		// проверяет ссылки на переменные внешнего цикла из вложенных функций.
		loopclosure.Analyzer,
		// проверяет неудачу вызова функции отмены контекста.
		lostcancel.Analyzer,
		// проверяет бесполезные сравнения с nil.
		nilfunc.Analyzer,
		// исследует граф управления потоком функции SSA и сообщает об ошибках, таких как разыменование nil-указателя и вырожденные сравнения с nil.
		nilness.Analyzer,
		// проверяет согласованность строк формата Printf и аргументов.
		printf.Analyzer,
		// проверяет случайное использование == или reflect.DeepEqual для сравнения значений reflect.Value.
		reflectvaluecompare.Analyzer,
		// проверяет сдвиги, превышающие ширину целого числа.
		shift.Analyzer,
		// обнаруживает неправильное использование не буферизированного сигнала в качестве аргумента для signal.Notify.
		sigchanyzer.Analyzer,
		// проверяет несоответствие пар ключ-значение в вызовах log/slog.
		slog.Analyzer,
		// проверяет вызовы sort.Slice, которые не используют тип среза в качестве первого аргумента.
		sortslice.Analyzer,
		// проверяет орфографические ошибки в подписях методов, похожих на известные интерфейсы.
		stdmethods.Analyzer,
		// выделяет преобразования типов из целых чисел в строки.
		stringintconv.Analyzer,
		// проверяет, что теги полей структуры правильно сформированы.
		structtag.Analyzer,
		// обнаруживает вызовов Fatal из тестовой горутины.
		testinggoroutine.Analyzer,
		// проверяет распространенные ошибочные использования тестов и примеров.
		tests.Analyzer,
		// проверяет использование вызовов time.Format или time.Parse с плохим форматом.
		timeformat.Analyzer,
		// проверяет передачу типов, не являющихся указателями или интерфейсами, функциям unmarshal и decode.
		unmarshal.Analyzer,
		// проверяет недостижимый код.
		unreachable.Analyzer,
		// проверяет недопустимые преобразования uintptr в unsafe.Pointer.
		unsafeptr.Analyzer,
		// проверяет неиспользованные результаты вызовов определенных чистых функций.
		unusedresult.Analyzer,
		// проверяет неиспользованные записи в элементы структуры или массива.
		unusedwrite.Analyzer,
		// проверяет использование обобщенных функций, добавленных в Go 1.18.
		usesgenerics.Analyzer,
		// проверяет, есть ли в функции main os.Exit
		customAnalysis.ErrNoExitAnalizer,
	}
	config := NewConfig()
	checks := make(map[string]bool, len(config.Staticcheck))
	for _, v := range config.Staticcheck {
		checks[v] = true
	}
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			staticChecks = append(staticChecks, v.Analyzer)
		}
	}
	multichecker.Main(staticChecks...)
}
