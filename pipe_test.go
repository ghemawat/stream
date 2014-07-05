package pipe

import (
	"fmt"
	_ "testing"
	"time"
)

func Example() {
	Print(
		Find(FILES, "."),
		Grep(`pipe.*\.go$`),
		NumberLines(),
	)
	// Output:
	//     1 pipe.go
	//     2 pipe_test.go
}

func ExampleEmpty() {
	Print()
	// Output:
}

func ExampleSingle() {
	Print(Echo("foo"))
	// Output:
	// foo
}

func ExampleMultiple() {
	Print(Echo("foo"), Echo("bar"))
	// Output:
	// foo
	// bar
}

func ExampleSequence_empty() {
	Print(Sequence())
	// Output:
}

func ExampleSequence_multi() {
	Print(Sequence(Echo("1 of 2"), Echo("2 of 2")))
	// Output:
	// 1 of 2
	// 2 of 2
}

func ExampleSequence_single() {
	Print(Sequence(Echo("1 of 1")))
	// Output:
	// 1 of 1
}

func ExampleForEach() {
	for s := range ForEach(Numbers(1, 10), First(3)) {
		fmt.Print(s)
	}
	// Output:
	// 123
}

func ExamplePrint() {
	Print(Echo("a"), Echo("b"), Echo("c"))
	// Output:
	// a
	// b
	// c
}

func ExampleEcho() {
	Print(Echo("hello", "world"))
	// Output:
	// hello
	// world
}

func ExampleNumbers() {
	Print(Numbers(2, 5))
	// Output:
	// 2
	// 3
	// 4
	// 5
}

func ExampleMap() {
	Print(
		Echo("hello", "there", "how", "are", "you?"),
		Map(func(s string) string {
			return fmt.Sprintf("%d %s", len(s), s)
		}),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleIf() {
	Print(Numbers(1, 12), If(func(s string) bool { return len(s) > 1 }))
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrep() {
	Print(Numbers(1, 12), Grep(".."))
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrepNot() {
	Print(Numbers(1, 12), GrepNot("^.$"))
	// Output:
	// 10
	// 11
	// 12
}

func ExampleUniq() {
	Print(Echo("a", "b", "b", "c", "b"), Uniq())
	// Output:
	// a
	// b
	// c
	// b
}

func ExampleUniqWithCount() {
	Print(
		Echo("a", "b", "b", "c"),
		UniqWithCount(),
	)
	// Output:
	// 1 a
	// 2 b
	// 1 c
}

func ExampleParallelMap() {
	Print(
		Echo("hello", "there", "how", "are", "you?"),
		ParallelMap(4, func(s string) string {
			// Sleep some amount to ensure that ParalellMap
			// implementation handles out of order results.
			time.Sleep(10 * time.Duration(len(s)) * time.Millisecond)
			return fmt.Sprintf("%d %s", len(s), s)
		}),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleSubstitute() {
	Print(Numbers(1, 5), Substitute("(3)", "$1$1"))
	// Output:
	// 1
	// 2
	// 33
	// 4
	// 5
}

func ExampleNumeric() {
	Print(
		Echo(
			"a 100",
			"b 20",
			"c notanumber", // Will sort last since column 2 is not a number
			"d",            // Will sort earliest since column 2 is missing
		),
		Sort(Numeric(2)),
	)
	// Output:
	// d
	// b 20
	// a 100
	// c notanumber
}

func ExampleTextual() {
	Print(
		Echo(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		Sort(Textual(2)),
	)
	// Output:
	// 30
	// 20 apples
	// 10 bananas
}

func ExampleDescending() {
	Print(
		Echo(
			"100",
			"20",
			"50",
		),
		Sort(Descending(Numeric(1))),
	)
	// Output:
	// 100
	// 50
	// 20
}

func ExampleSort() {
	Print(Echo("banana", "apple", "cheese", "apple"), Sort())
	// Output:
	// apple
	// apple
	// banana
	// cheese
}
func ExampleSort_twoTextColumns() {
	Print(
		Echo(
			"2 green bananas",
			"3 red apples",
			"4 yellow bananas",
			"5 brown pears",
			"6 green apples",
		),
		Sort(Textual(2), Textual(3)),
	)
	// Output:
	// 5 brown pears
	// 6 green apples
	// 2 green bananas
	// 3 red apples
	// 4 yellow bananas
}

func ExampleSort_twoNumericColumns() {
	Print(
		Echo(
			"1970 12",
			"1970 6",
			"1950 6",
			"1980 9",
		),
		Sort(Numeric(1), Numeric(2)))
	// Output:
	// 1950 6
	// 1970 6
	// 1970 12
	// 1980 9
}

func ExampleSort_mixedColumns() {
	Print(
		Echo(
			"1970 march",
			"1970 feb",
			"1950 june",
			"1980 sep",
		),
		Sort(Numeric(1), Textual(2)))
	// Output:
	// 1950 june
	// 1970 feb
	// 1970 march
	// 1980 sep
}

func ExampleReverse() {
	Print(Echo("a", "b"), Reverse())
	// Output:
	// b
	// a
}

func ExampleSample() {
	Print(
		Numbers(100, 200),
		Sample(4),
		Sort(),
	)
	// Output:
	// 122
	// 150
	// 154
	// 158
}

func ExampleFirst() {
	Print(Numbers(1, 10), First(3))
	// Output:
	// 1
	// 2
	// 3
}

func ExampleLast() {
	Print(Numbers(1, 10), Last(2))
	// Output:
	// 9
	// 10
}

func ExampleDropFirst() {
	Print(Numbers(1, 10), DropFirst(8))
	// Output:
	// 9
	// 10
}

func ExampleDropLast() {
	Print(Numbers(1, 10), DropLast(3))
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
}

func ExampleNumberLines() {
	Print(Echo("a", "b"), NumberLines())
	// Output:
	//     1 a
	//     2 b
}

func ExampleCut() {
	Print(Echo("hello", "world."), Cut(2, 4))
	// Output:
	// llo
	// rld
}

func ExmapleSelect() {
	Print(Echo("hello world"), Select(2, 0, 1))
	// Output:
	// world hello world hello
}

func ExampleFind() {
	Print(Find(FILES, "."), Grep("pipe"))
	// Output:
	// pipe.go
	// pipe_test.go
}

func ExampleFind_dirs() {
	Print(Find(DIRS, "."), GrepNot("git"))
	// Output:
	// .
}

func ExampleCat() {
	Print(Cat("pipe_test.go"), Grep("^func ExampleCat"))
	// Output:
	// func ExampleCat() {
}

func ExampleCommandOutput() {
	Print(
		CommandOutput("find", ".", "-type", "f", "-print"),
		Grep(`^\./pipe.*\.go$`),
		Sort(),
	)

	// Output:
	// ./pipe.go
	// ./pipe_test.go
}
