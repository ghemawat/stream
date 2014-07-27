package pipe_test

import (
	"github.com/ghemawat/pipe"

	"bytes"
	"fmt"
	"os"
)

func ExampleSequence() {
	pipe.ForEach(pipe.Sequence(
		pipe.Numbers(1, 25),
		pipe.Grep("3"),
	), func(s string) { fmt.Println(s) })
	// Output:
	// 3
	// 13
	// 23
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

func ExampleOutput() {
	out, err := pipe.Output(pipe.Numbers(1, 3))
	fmt.Println(out, err)
	// Output:
	// [1 2 3] <nil>
}

func ExampleRun() {
	pipe.Run(
		pipe.Items("line 1", "line 2"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// line 1
	// line 2
}

func ExampleItems() {
	pipe.Run(
		pipe.Items("hello", "world"),
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
		pipe.Items("hello", "there", "how", "are", "you?"),
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
		pipe.Items("a", "b", "b", "c", "b"),
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
		pipe.Items("a", "b", "b", "c"),
		pipe.UniqWithCount(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1 a
	// 2 b
	// 1 c
}

func ExampleParallel() {
	pipe.Run(
		pipe.Items("hello", "there", "how", "are", "you?"),
		pipe.Parallel(4,
			pipe.Map(func(s string) string {
				return fmt.Sprintf("%d %s", len(s), s)
			}),
		),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 3 are
	// 3 how
	// 4 you?
	// 5 hello
	// 5 there
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

func ExampleSort() {
	pipe.Run(
		pipe.Items("banana", "apple", "cheese", "apple"),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// apple
	// apple
	// banana
	// cheese
}

func ExampleSort_multipleColumns() {
	// Sort numerically by column 1. Break ties by sorting
	// lexicographically by column 2.
	pipe.Run(
		pipe.Items(
			"1970 march",
			"1970 feb",
			"1950 june",
			"1980 sep",
		),
		pipe.Sort().Num(1).Text(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1950 june
	// 1970 feb
	// 1970 march
	// 1980 sep
}

func ExampleSorter_Num() {
	pipe.Run(
		pipe.Items(
			"a 100",
			"b 20",
			"c notanumber", // Will sort last since column 2 is not a number
			"d",            // Will sort earliest since column 2 is missing
		),
		pipe.Sort().Num(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// d
	// b 20
	// a 100
	// c notanumber
}

func ExampleSorter_NumDecreasing() {
	pipe.Run(
		pipe.Items(
			"a 100",
			"b 20",
			"c notanumber",
			"d",
		),
		pipe.Sort().NumDecreasing(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// c notanumber
	// a 100
	// b 20
	// d
}

func ExampleSorter_Text() {
	pipe.Run(
		pipe.Items(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		pipe.Sort().Text(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 30
	// 20 apples
	// 10 bananas
}

func ExampleSorter_TextDecreasing() {
	pipe.Run(
		pipe.Items(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		pipe.Sort().TextDecreasing(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 10 bananas
	// 20 apples
	// 30
}

func ExampleSorter_By() {
	pipe.Run(
		pipe.Items("bananas", "apples", "pears"),
		pipe.Sort().By(func(a, b string) bool { return len(a) < len(b) }),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// pears
	// apples
	// bananas
}

func ExampleReverse() {
	pipe.Run(
		pipe.Items("a", "b"),
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
		pipe.WriteLines(os.Stdout),
	)
	// Output not checked since it is non-deterministic.
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
		pipe.Items("a", "b"),
		pipe.NumberLines(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	//     1 a
	//     2 b
}

func ExampleSlice() {
	pipe.Run(
		pipe.Items("hello", "world."),
		pipe.Slice(2, 5),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// llo
	// rld
}

func ExampleColumns() {
	pipe.Run(
		pipe.Items("hello world"),
		pipe.Columns(2, 3, 1),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// world hello
}

func ExampleFind() {
	pipe.Run(
		pipe.Find(".").Files(),
		pipe.Grep("pipe"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// pipe.go
	// pipe_test.go
}

func ExampleFindFilter_SkipDir() {
	pipe.Run(
		pipe.Find(".").SkipDir(".git"),
		pipe.Grep("x"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// regexp.go
	// xargs.go
}

func ExampleFind_error() {
	err := pipe.Run(pipe.Find("/no_such_dir"))
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

func ExampleCommand() {
	pipe.Run(
		pipe.Numbers(1, 100),
		pipe.Command("wc", "-l"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 100
}

func ExampleCommand_outputOnly() {
	pipe.Run(
		pipe.Command("find", ".", "-type", "f", "-print"),
		pipe.Grep(`^\./pipe.*\.go$`),
		pipe.Sort(),
		pipe.WriteLines(os.Stdout),
	)

	// Output:
	// ./pipe.go
	// ./pipe_test.go
}

func ExampleCommand_withError() {
	err := pipe.Run(pipe.Command("no_such_command"))
	if err == nil {
		fmt.Println("execution of missing command succeeded unexpectedly")
	}
	// Output:
}

func ExampleXargs() {
	pipe.Run(
		pipe.Numbers(1, 5),
		pipe.Xargs("echo"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1 2 3 4 5
}

func ExampleXargsFilter_LimitArgs() {
	pipe.Run(
		pipe.Numbers(1, 5),
		pipe.Xargs("echo").LimitArgs(2),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 1 2
	// 3 4
	// 5
}

func ExampleXargs_splitArguments() {
	// Xargs should split the long list of arguments into
	// three executions to keep command length below 4096.
	pipe.Run(
		pipe.Numbers(1, 2000),
		pipe.Xargs("echo"),
		pipe.Command("wc", "-l"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// 3
}
