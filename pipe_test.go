package pipe_test

import (
	"github.com/ghemawat/pipe"

	"bytes"
	"fmt"
	"os"
	"regexp"
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

func Example_error() {
	counter := pipe.FilterFunc(func(arg pipe.Arg) error {
		re, err := regexp.Compile("[")
		if err != nil {
			return err
		}
		n := 1
		for s := range arg.In {
			if re.MatchString(s) {
				n++
			}
		}
		arg.Out <- fmt.Sprint(n)
		return nil
	})
	err := pipe.Run(
		pipe.Numbers(1, 100),
		counter,
		pipe.WriteLines(os.Stdout),
	)
	if err == nil {
		fmt.Println("did not catch error")
	}
	// Output:
}

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
	result, err := pipe.Output(pipe.Numbers(1, 3))
	for _, s := range result {
		fmt.Println(s)
	}
	fmt.Println("error:", err)
	// Output:
	// 1
	// 2
	// 3
	// error: <nil>
}

func ExampleRun() {
	pipe.Run(
		pipe.Echo("line 1", "line 2"),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// line 1
	// line 2
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

func ExampleParallel() {
	pipe.Run(
		pipe.Echo("hello", "there", "how", "are", "you?"),
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
		pipe.Echo("a", "b"),
		pipe.NumberLines(),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	//     1 a
	//     2 b
}

func ExampleSlice() {
	pipe.Run(
		pipe.Echo("hello", "world."),
		pipe.Slice(2, 5),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// llo
	// rld
}

func ExampleColumns() {
	pipe.Run(
		pipe.Echo("hello world"),
		pipe.Columns(2, 3, 1),
		pipe.WriteLines(os.Stdout),
	)
	// Output:
	// world hello
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
	err := pipe.Run(
		pipe.Numbers(1, 5),
		pipe.Xargs("echo"),
		pipe.WriteLines(os.Stdout),
	)
	fmt.Println("error:", err)
	// Output:
	// 1 2 3 4 5
	// error: <nil>
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
