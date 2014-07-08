package pipe

import (
	"fmt"
	"os"
	_ "testing"
	"time"
)

func Example() {
	err := Run(
		Find(FILES, "."),
		Grep(`pipe.*\.go$`),
		Tee(os.Stdout),
	)
	fmt.Println("error:", err)
	// Output:
	// pipe.go
	// pipe_test.go
	// error: <nil>
}

func ExampleSequence() {
	ForEach(Sequence(
		Echo("1 of 3"),
		Echo("2 of 3"),
		Echo("3 of 3"),
	), func(s string) { fmt.Println(s) })
	// Output:
	// 1 of 3
	// 2 of 3
	// 3 of 3
}

func ExampleForEach() {
	err := ForEach(Numbers(1, 5), func(s string) {
		fmt.Print(s)
	})
	if err != nil {
		panic(err)
	}
	// Output:
	// 12345
}

func ExampleEcho() {
	Run(
		Echo("hello", "world"),
		Tee(os.Stdout),
	)
	// Output:
	// hello
	// world
}

func ExampleNumbers() {
	Run(
		Numbers(2, 5),
		Tee(os.Stdout),
	)
	// Output:
	// 2
	// 3
	// 4
	// 5
}

func ExampleMap() {
	Run(
		Echo("hello", "there", "how", "are", "you?"),
		Map(func(s string) string {
			return fmt.Sprintf("%d %s", len(s), s)
		}),
		Tee(os.Stdout),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleIf() {
	Run(
		Numbers(1, 12),
		If(func(s string) bool { return len(s) > 1 }),
		Tee(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrep() {
	Run(
		Numbers(1, 12),
		Grep(".."),
		Tee(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrepNot() {
	Run(
		Numbers(1, 12),
		GrepNot("^.$"),
		Tee(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleUniq() {
	Run(
		Echo("a", "b", "b", "c", "b"),
		Uniq(),
		Tee(os.Stdout),
	)
	// Output:
	// a
	// b
	// c
	// b
}

func ExampleUniqWithCount() {
	Run(
		Echo("a", "b", "b", "c"),
		UniqWithCount(),
		Tee(os.Stdout),
	)
	// Output:
	// 1 a
	// 2 b
	// 1 c
}

func ExampleParallelMap() {
	Run(
		Echo("hello", "there", "how", "are", "you?"),
		ParallelMap(4, func(s string) string {
			// Sleep some amount to ensure that ParalellMap
			// implementation handles out of order results.
			time.Sleep(10 * time.Duration(len(s)) * time.Millisecond)
			return fmt.Sprintf("%d %s", len(s), s)
		}),
		Tee(os.Stdout),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleSubstitute() {
	Run(
		Numbers(1, 5),
		Substitute("(3)", "$1$1"),
		Tee(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 33
	// 4
	// 5
}

func ExampleNumeric() {
	Run(
		Echo(
			"a 100",
			"b 20",
			"c notanumber", // Will sort last since column 2 is not a number
			"d",            // Will sort earliest since column 2 is missing
		),
		Sort(Numeric(2)),
		Tee(os.Stdout),
	)
	// Output:
	// d
	// b 20
	// a 100
	// c notanumber
}

func ExampleTextual() {
	Run(
		Echo(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		Sort(Textual(2)),
		Tee(os.Stdout),
	)
	// Output:
	// 30
	// 20 apples
	// 10 bananas
}

func ExampleDescending() {
	Run(
		Echo(
			"100",
			"20",
			"50",
		),
		Sort(Descending(Numeric(1))),
		Tee(os.Stdout),
	)
	// Output:
	// 100
	// 50
	// 20
}

func ExampleSort() {
	Run(
		Echo("banana", "apple", "cheese", "apple"),
		Sort(),
		Tee(os.Stdout),
	)
	// Output:
	// apple
	// apple
	// banana
	// cheese
}
func ExampleSort_twoTextColumns() {
	Run(
		Echo(
			"2 green bananas",
			"3 red apples",
			"4 yellow bananas",
			"5 brown pears",
			"6 green apples",
		),
		Sort(Textual(2), Textual(3)),
		Tee(os.Stdout),
	)
	// Output:
	// 5 brown pears
	// 6 green apples
	// 2 green bananas
	// 3 red apples
	// 4 yellow bananas
}

func ExampleSort_twoNumericColumns() {
	Run(
		Echo(
			"1970 12",
			"1970 6",
			"1950 6",
			"1980 9",
		),
		Sort(Numeric(1), Numeric(2)),
		Tee(os.Stdout),
	)
	// Output:
	// 1950 6
	// 1970 6
	// 1970 12
	// 1980 9
}

func ExampleSort_mixedColumns() {
	Run(
		Echo(
			"1970 march",
			"1970 feb",
			"1950 june",
			"1980 sep",
		),
		Sort(Numeric(1), Textual(2)),
		Tee(os.Stdout),
	)
	// Output:
	// 1950 june
	// 1970 feb
	// 1970 march
	// 1980 sep
}

func ExampleReverse() {
	Run(
		Echo("a", "b"),
		Reverse(),
		Tee(os.Stdout),
	)
	// Output:
	// b
	// a
}

func ExampleSample() {
	Run(
		Numbers(100, 200),
		Sample(4),
		Sort(),
		Tee(os.Stdout),
	)
	// Output:
	// 122
	// 150
	// 154
	// 158
}

func ExampleFirst() {
	Run(
		Numbers(1, 10),
		First(3),
		Tee(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 3
}

func ExampleLast() {
	Run(
		Numbers(1, 10),
		Last(2),
		Tee(os.Stdout),
	)
	// Output:
	// 9
	// 10
}

func ExampleDropFirst() {
	Run(
		Numbers(1, 10),
		DropFirst(8),
		Tee(os.Stdout),
	)
	// Output:
	// 9
	// 10
}

func ExampleDropLast() {
	Run(
		Numbers(1, 10),
		DropLast(3),
		Tee(os.Stdout),
	)
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
	Run(
		Echo("a", "b"),
		NumberLines(),
		Tee(os.Stdout),
	)
	// Output:
	//     1 a
	//     2 b
}

func ExampleCut() {
	Run(
		Echo("hello", "world."),
		Cut(2, 4),
		Tee(os.Stdout),
	)
	// Output:
	// llo
	// rld
}

func ExmapleSelect() {
	Run(
		Echo("hello world"),
		Select(2, 3, 0, 1),
		Tee(os.Stdout),
	)
	// Output:
	// world hello world hello
}

func ExampleFind() {
	Run(
		Find(FILES, "."),
		Grep("pipe"),
		Tee(os.Stdout),
	)
	// Output:
	// pipe.go
	// pipe_test.go
}

func ExampleFind_dirs() {
	Run(
		Find(DIRS, "."),
		GrepNot("git"),
		Tee(os.Stdout),
	)
	// Output:
	// .
}

func ExampleFind_error() {
	if ForEach(Find(ALL, "/no_such_dir"), func(string) {}) == nil {
		fmt.Println("Find did not return expected error")
	}
	// Output:
}

func ExampleCat() {
	Run(
		Cat("pipe_test.go"),
		Grep("^func ExampleCat"),
		Tee(os.Stdout),
	)
	// Output:
	// func ExampleCat() {
}

func ExampleCommandOutput() {
	Run(
		CommandOutput("find", ".", "-type", "f", "-print"),
		Grep(`^\./pipe.*\.go$`),
		Sort(),
		Tee(os.Stdout),
	)

	// Output:
	// ./pipe.go
	// ./pipe_test.go
}

func ExampleCommandOutput_error() {
	if ForEach(CommandOutput("no_such_command"), func(string) {}) == nil {
		fmt.Println("execution of missing command succeeded unexpectedly")
	}
	// Output:
}
