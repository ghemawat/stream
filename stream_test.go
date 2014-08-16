package stream_test

import (
	"github.com/ghemawat/stream"

	"bytes"
	"fmt"
	"os"
)

func ExampleSequence() {
	stream.ForEach(stream.Sequence(
		stream.Numbers(1, 25),
		stream.Grep("3"),
	), func(s string) { fmt.Println(s) })
	// Output:
	// 3
	// 13
	// 23
}

func ExampleForEach() {
	err := stream.ForEach(stream.Numbers(1, 5), func(s string) {
		fmt.Print(s)
	})
	if err != nil {
		panic(err)
	}
	// Output:
	// 12345
}

func ExampleContents() {
	out, err := stream.Contents(stream.Numbers(1, 3))
	fmt.Println(out, err)
	// Output:
	// [1 2 3] <nil>
}

func ExampleRun() {
	err := stream.Run(
		stream.Items("line 1", "line 2"),
		stream.WriteLines(os.Stdout),
	)
	fmt.Println("error:", err)
	// Output:
	// line 1
	// line 2
	// error: <nil>
}

func ExampleItems() {
	stream.Run(
		stream.Items("hello", "world"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// hello
	// world
}

func ExampleNumbers() {
	stream.Run(
		stream.Numbers(2, 5),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 2
	// 3
	// 4
	// 5
}

func ExampleMap() {
	stream.Run(
		stream.Items("hello", "there", "how", "are", "you?"),
		stream.Map(func(s string) string {
			return fmt.Sprintf("%d %s", len(s), s)
		}),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 5 hello
	// 5 there
	// 3 how
	// 3 are
	// 4 you?
}

func ExampleIf() {
	stream.Run(
		stream.Numbers(1, 12),
		stream.If(func(s string) bool { return len(s) > 1 }),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrep() {
	stream.Run(
		stream.Numbers(1, 12),
		stream.Grep(".."),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleGrepNot() {
	stream.Run(
		stream.Numbers(1, 12),
		stream.GrepNot("^.$"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 10
	// 11
	// 12
}

func ExampleUniq() {
	stream.Run(
		stream.Items("a", "b", "b", "c", "b"),
		stream.Uniq(),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// a
	// b
	// c
	// b
}

func ExampleUniqWithCount() {
	stream.Run(
		stream.Items("a", "b", "b", "c"),
		stream.UniqWithCount(),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1 a
	// 2 b
	// 1 c
}

func ExampleParallel() {
	stream.Run(
		stream.Items("hello", "there", "how", "are", "you?"),
		stream.Parallel(4,
			stream.Map(func(s string) string {
				return fmt.Sprintf("%d %s", len(s), s)
			}),
		),
		stream.Sort(),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 3 are
	// 3 how
	// 4 you?
	// 5 hello
	// 5 there
}

func ExampleSubstitute() {
	stream.Run(
		stream.Numbers(1, 5),
		stream.Substitute("(3)", "$1$1"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 33
	// 4
	// 5
}

func ExampleSort() {
	stream.Run(
		stream.Items("banana", "apple", "cheese", "apple"),
		stream.Sort(),
		stream.WriteLines(os.Stdout),
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
	stream.Run(
		stream.Items(
			"1970 march",
			"1970 feb",
			"1950 june",
			"1980 sep",
		),
		stream.Sort().Num(1).Text(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1950 june
	// 1970 feb
	// 1970 march
	// 1980 sep
}

func ExampleSorter_Num() {
	stream.Run(
		stream.Items(
			"a 100",
			"b 20.3",
			"c notanumber", // Will sort last since column 2 is not a number
			"d",            // Will sort earliest since column 2 is missing
		),
		stream.Sort().Num(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// d
	// b 20.3
	// a 100
	// c notanumber
}

func ExampleSorter_NumDecreasing() {
	stream.Run(
		stream.Items(
			"a 100",
			"b 20",
			"c notanumber",
			"d",
		),
		stream.Sort().NumDecreasing(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// c notanumber
	// a 100
	// b 20
	// d
}

func ExampleSorter_Text() {
	stream.Run(
		stream.Items(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		stream.Sort().Text(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 30
	// 20 apples
	// 10 bananas
}

func ExampleSorter_TextDecreasing() {
	stream.Run(
		stream.Items(
			"10 bananas",
			"20 apples",
			"30", // Will sort first since column 2 is missing
		),
		stream.Sort().TextDecreasing(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 10 bananas
	// 20 apples
	// 30
}

func ExampleSorter_By() {
	stream.Run(
		stream.Items("bananas", "apples", "pears"),
		stream.Sort().By(func(a, b string) bool { return len(a) < len(b) }),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// pears
	// apples
	// bananas
}

func ExampleReverse() {
	stream.Run(
		stream.Items("a", "b"),
		stream.Reverse(),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// b
	// a
}

func ExampleSample() {
	stream.Run(
		stream.Numbers(100, 200),
		stream.Sample(4),
		stream.WriteLines(os.Stdout),
	)
	// Output not checked since it is non-deterministic.
}

func ExampleSampleWithSeed() {
	stream.Run(
		stream.Numbers(1, 100),
		stream.SampleWithSeed(2, 100),
		stream.Sort().Num(1),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 11
	// 46
}

func ExampleFirst() {
	stream.Run(
		stream.Numbers(1, 10),
		stream.First(3),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 3
}

func ExampleLast() {
	stream.Run(
		stream.Numbers(1, 10),
		stream.Last(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 9
	// 10
}

func ExampleDropFirst() {
	stream.Run(
		stream.Numbers(1, 10),
		stream.DropFirst(8),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 9
	// 10
}

func ExampleDropLast() {
	stream.Run(
		stream.Numbers(1, 10),
		stream.DropLast(3),
		stream.WriteLines(os.Stdout),
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
	stream.Run(
		stream.Items("a", "b"),
		stream.NumberLines(),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	//     1 a
	//     2 b
}

func ExampleColumns() {
	stream.Run(
		stream.Items("hello world"),
		stream.Columns(2, 3, 1),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// world hello
}

func ExampleFind() {
	stream.Run(
		stream.Find(".").IfMode(os.FileMode.IsRegular),
		stream.Grep("stream"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// stream.go
	// stream_test.go
}

func ExampleFindFilter_SkipDirIf() {
	stream.Run(
		stream.Find(".").SkipDirIf(func(d string) bool { return d == ".git" }),
		stream.Grep("x"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// regexp.go
	// xargs.go
}

func ExampleFind_error() {
	err := stream.Run(stream.Find("/no_such_dir"))
	if err == nil {
		fmt.Println("stream.Find did not return expected error")
	}
	// Output:
}

func ExampleCat() {
	stream.Run(
		stream.Cat("stream_test.go"),
		stream.Grep("^func ExampleCat"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// func ExampleCat() {
}

func ExampleWriteLines() {
	stream.Run(
		stream.Numbers(1, 3),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1
	// 2
	// 3
}

func ExampleReadLines() {
	stream.Run(
		stream.ReadLines(bytes.NewBufferString("the\nquick\nbrown\nfox\n")),
		stream.Sort(),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// brown
	// fox
	// quick
	// the
}

func ExampleCommand() {
	stream.Run(
		stream.Numbers(1, 100),
		stream.Command("wc", "-l"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 100
}

func ExampleCommand_outputOnly() {
	stream.Run(
		stream.Command("find", ".", "-type", "f", "-print"),
		stream.Grep(`^\./stream.*\.go$`),
		stream.Sort(),
		stream.WriteLines(os.Stdout),
	)

	// Output:
	// ./stream.go
	// ./stream_test.go
}

func ExampleCommand_withError() {
	err := stream.Run(stream.Command("no_such_command"))
	if err == nil {
		fmt.Println("execution of missing command succeeded unexpectedly")
	}
	// Output:
}

func ExampleXargs() {
	stream.Run(
		stream.Numbers(1, 5),
		stream.Xargs("echo"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1 2 3 4 5
}

func ExampleXargsFilter_LimitArgs() {
	stream.Run(
		stream.Numbers(1, 5),
		stream.Xargs("echo").LimitArgs(2),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 1 2
	// 3 4
	// 5
}

func ExampleXargs_splitArguments() {
	// Xargs should split the long list of arguments into
	// three executions to keep command length below 4096.
	stream.Run(
		stream.Numbers(1, 2000),
		stream.Xargs("echo"),
		stream.Command("wc", "-l"),
		stream.WriteLines(os.Stdout),
	)
	// Output:
	// 3
}
