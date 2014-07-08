package pipe_test

import (
	"bytes"
	"fmt"
	"os"
	"pipe"
	_ "testing"
	"time"
)

func Example() {
	err := pipe.Run(
		pipe.Find(pipe.FILES, "."),
		pipe.Grep(`pipe.*\.go$`),
		pipe.WriteLines(os.Stdout),
	)
	fmt.Println("error:", err)
	// Output:
	// pipe.go
	// pipe_test.go
	// error: <nil>
}

func ExampleSequence() {
	pipe.ForEach(pipe.Sequence(
		pipe.Echo("1 of 3"),
		pipe.Echo("2 of 3"),
		pipe.Echo("3 of 3"),
	), func(s string) { fmt.Println(s) })
	// Output:
	// 1 of 3
	// 2 of 3
	// 3 of 3
}

func ExampleForEach() {
	err := pipe.ForEach(pipe.Numbers(1, 5), func(s string) {
		fmt.Print(s)
	})
	if err != nil {
		panic(err)
	}
	// Output:
	// 12345
}

func ExampleRun() {
	pipe.Run(
		pipe.Echo("line 1"),
		pipe.Echo("line 2"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// line 1
	// line 2
	// line 3
}

func ExampleEcho() {
	pipe.Run(
		pipe.Echo("hello", "world"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// hello
	// world
}

func ExampleNumbers() {
	pipe.Run(
		pipe.Numbers(2, 5),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 2
	// 3
	// 4
	// 5
}

func ExampleMap() {
	pipe.Run(
		pipe.Echo("hello", "there", "how", "are", "you?"),
		pipe.Map(func(s string) string {
			return fmt.Sprintf("%d %s", len(s), s)
		}),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleIf() {
	pipe.Run(
		pipe.Numbers(1, 12),
		pipe.If(func(s string) bool { return len(s) > 1 }),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrep() {
	pipe.Run(
		pipe.Numbers(1, 12),
		pipe.Grep(".."),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrepNot() {
	pipe.Run(
		pipe.Numbers(1, 12),
		pipe.GrepNot("^.$"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleUniq() {
	pipe.Run(
		pipe.Echo("a", "b", "b", "c", "b"),
		pipe.Uniq(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// a
	// b
	// c
	// b
}

func ExampleUniqWithCount() {
	pipe.Run(
		pipe.Echo("a", "b", "b", "c"),
		pipe.UniqWithCount(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1 a
	// 2 b
	// 1 c
}

func ExampleParallelMap() {
	pipe.Run(
		pipe.Echo("hello", "there", "how", "are", "you?"),
		pipe.ParallelMap(4, func(s string) string {
			// Sleep some amount to ensure that ParalellMap
			// implementation handles out of order results.
			time.Sleep(10 * time.Duration(len(s)) * time.Millisecond)
			return fmt.Sprintf("%d %s", len(s), s)
		}),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleSubstitute() {
	pipe.Run(
		pipe.Numbers(1, 5),
		pipe.Substitute("(3)", "$1$1"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 33
	// 4
	// 5
}

func ExampleNumeric() {
	pipe.Run(
		pipe.Echo(
			"a 100",
			"b 20",
			"c notanumber", // Will sort last since column 2 is not a number
			"d",            // Will sort earliest since column 2 is missing
		),
		pipe.Sort(pipe.Numeric(2)),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// d
	// b 20
	// a 100
	// c notanumber
}

func ExampleTextual() {
	pipe.Run(
		pipe.Echo(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		pipe.Sort(pipe.Textual(2)),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 30
	// 20 apples
	// 10 bananas
}

func ExampleDescending() {
	pipe.Run(
		pipe.Echo(
			"100",
			"20",
			"50",
		),
		pipe.Sort(pipe.Descending(pipe.Numeric(1))),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 100
	// 50
	// 20
}

func ExampleSort() {
	pipe.Run(
		pipe.Echo("banana", "apple", "cheese", "apple"),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// apple
	// apple
	// banana
	// cheese
}
func ExampleSort_twoTextColumns() {
	pipe.Run(
		pipe.Echo(
			"2 green bananas",
			"3 red apples",
			"4 yellow bananas",
			"5 brown pears",
			"6 green apples",
		),
		pipe.Sort(pipe.Textual(2), pipe.Textual(3)),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 5 brown pears
	// 6 green apples
	// 2 green bananas
	// 3 red apples
	// 4 yellow bananas
}

func ExampleSort_twoNumericColumns() {
	pipe.Run(
		pipe.Echo(
			"1970 12",
			"1970 6",
			"1950 6",
			"1980 9",
		),
		pipe.Sort(pipe.Numeric(1), pipe.Numeric(2)),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1950 6
	// 1970 6
	// 1970 12
	// 1980 9
}

func ExampleSort_mixedColumns() {
	pipe.Run(
		pipe.Echo(
			"1970 march",
			"1970 feb",
			"1950 june",
			"1980 sep",
		),
		pipe.Sort(pipe.Numeric(1), pipe.Textual(2)),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1950 june
	// 1970 feb
	// 1970 march
	// 1980 sep
}

func ExampleReverse() {
	pipe.Run(
		pipe.Echo("a", "b"),
		pipe.Reverse(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// b
	// a
}

func ExampleSample() {
	pipe.Run(
		pipe.Numbers(100, 200),
		pipe.Sample(4),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 122
	// 150
	// 154
	// 158
}

func ExampleFirst() {
	pipe.Run(
		pipe.Numbers(1, 10),
		pipe.First(3),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 3
}

func ExampleLast() {
	pipe.Run(
		pipe.Numbers(1, 10),
		pipe.Last(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 9
	// 10
}

func ExampleDropFirst() {
	pipe.Run(
		pipe.Numbers(1, 10),
		pipe.DropFirst(8),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 9
	// 10
}

func ExampleDropLast() {
	pipe.Run(
		pipe.Numbers(1, 10),
		pipe.DropLast(3),
		pipe.WriteLines(os.Stdout),
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
	pipe.Run(
		pipe.Echo("a", "b"),
		pipe.NumberLines(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	//     1 a
	//     2 b
}

func ExampleCut() {
	pipe.Run(
		pipe.Echo("hello", "world."),
		pipe.Cut(2, 4),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// llo
	// rld
}

func ExmapleSelect() {
	pipe.Run(
		pipe.Echo("hello world"),
		pipe.Select(2, 3, 0, 1),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// world hello world hello
}

func ExampleFind() {
	pipe.Run(
		pipe.Find(pipe.FILES, "."),
		pipe.Grep("pipe"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// pipe.go
	// pipe_test.go
}

func ExampleFind_dirs() {
	pipe.Run(
		pipe.Find(pipe.DIRS, "."),
		pipe.GrepNot("git"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// .
}

func ExampleFind_error() {
	err := pipe.Run(pipe.Find(pipe.ALL, "/no_such_dir"))
	if err == nil {
		fmt.Println("pipe.Find did not return expected error")
	}
	// Output:
}

func ExampleCat() {
	pipe.Run(
		pipe.Cat("pipe_test.go"),
		pipe.Grep("^func ExampleCat"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// func ExampleCat() {
}

func ExampleWriteLines() {
	pipe.Run(
		pipe.Numbers(1, 3),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 3
}

func ExampleReadLines() {
	pipe.Run(
		pipe.ReadLines(bytes.NewBufferString("the\nquick\nbrown\nfox\n")),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// brown
	// fox
	// quick
	// the
}

func ExampleCommandOutput() {
	pipe.Run(
		pipe.CommandOutput("find", ".", "-type", "f", "-print"),
		pipe.Grep(`^\./pipe.*\.go$`),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)

	// Output:
	// ./pipe.go
	// ./pipe_test.go
}

func ExampleCommandOutput_error() {
	err := pipe.Run(pipe.CommandOutput("no_such_command"))
	if err == nil {
		fmt.Println("execution of missing command succeeded unexpectedly")
	}
	// Output:
}
