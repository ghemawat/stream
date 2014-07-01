package pipe

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	_ "testing"
)

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

func ExampleSequence() {
	Print(Sequence())
	fmt.Println("---")
	Print(Sequence(Echo("1 of 1")))
	Print(Sequence(Echo("1 of 2"), Echo("2 of 2")))
	// Output:
	// ---
	// 1 of 1
	// 1 of 2
	// 2 of 2
}

func ExampleForEach() {
	for s := range ForEach(Numbers(2, 4)) {
		fmt.Print(s)
	}
	// Output:
	// 234
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
	Print(Echo("a", "b", "b", "c"), Uniq)
	// Output:
	// a
	// b
	// c
}

func ExampleUniqWithCount() {
	Print(Echo("a", "b", "b", "c"), UniqWithCount)
	// Output:
	// 1 a
	// 2 b
	// 1 c
}

func ExampleParallel() {
	Print(
		Numbers(1, 3),
		Parallel(4, func(s string, out chan<- string) { out <- s }),
		Sort(),
	)
	// Output:
	// 1
	// 2
	// 3
}

func ExampleReplaceMatch() {
	Print(Numbers(1, 5), ReplaceMatch("(3)", "$1$1"))
	// Output:
	// 1
	// 2
	// 33
	// 4
	// 5
}

func ExampleDeleteMatch() {
	Print(Numbers(1, 5), DeleteMatch("[24]"))
	// Output:
	// 1
	//
	// 3
	//
	// 5
}

func sortData() Filter {
	return Echo(
		"8 1",
		"8 3 x",
		"8 3 w",
		"8 2",
		"4 5",
		"9 3",
		"12 13",
		"12 5",
	)
}

func ExampleSort() {
	Print(sortData(), Sort())
	// Output:
	// 12 13
	// 12 5
	// 4 5
	// 8 1
	// 8 2
	// 8 3 w
	// 8 3 x
	// 9 3
}

func ExampleSort_TextCol() {
	Print(sortData(), Sort(Text(2)))
	// Output:
	// 8 1
	// 12 13
	// 8 2
	// 8 3 w
	// 8 3 x
	// 9 3
	// 12 5
	// 4 5
}

func ExampleSort_TwoText() {
	Print(sortData(), Sort(Text(1), Text(2)))
	// Output:
	// 12 13
	// 12 5
	// 4 5
	// 8 1
	// 8 2
	// 8 3 w
	// 8 3 x
	// 9 3
}

func ExampleSort_TwoNum() {
	Print(sortData(), Sort(Num(1), Num(2)))
	// Output:
	// 4 5
	// 8 1
	// 8 2
	// 8 3 w
	// 8 3 x
	// 9 3
	// 12 5
	// 12 13
}

func ExampleSort_Mix() {
	Print(sortData(), Sort(Text(1), Num(2)))
	// Output:
	// 12 5
	// 12 13
	// 4 5
	// 8 1
	// 8 2
	// 8 3 w
	// 8 3 x
	// 9 3
}

func ExampleSort_Rev() {
	Print(sortData(), Sort(Rev(Num(1)), Num(2)))
	// Output:
	// 12 5
	// 12 13
	// 9 3
	// 8 1
	// 8 2
	// 8 3 w
	// 8 3 x
	// 4 5
}

func ExampleReverse() {
	Print(Echo("a", "b"), Reverse)
	// Output:
	// b
	// a
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
	Print(Numbers(1, 10), DropLast(8))
	// Output:
	// 1
	// 2
}

func ExampleNumberLines() {
	Print(Echo("a", "b"), NumberLines)
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
	Print(Find(DIRS, "."), GrepNot("git"), Echo("---"))
	Print(Find(FILES, "."), Grep("pipe"))
	// Output:
	// .
	// ---
	// pipe.go
	// pipe_test.go
}

func ExampleCat() {
	Print(Cat("pipe_test.go"), Grep("^func ExampleCat"))
	// Output:
	// func ExampleCat() {
}

func ExampleSystem() {
	Print(
		System("find", ".", "-type", "f", "-print"),
		Grep(`^\./pipe.*\.go$`),
		Sort(),
	)

	// Output:
	// ./pipe.go
	// ./pipe_test.go
}

func ExampleMix() {
	dbl := func(arg Arg) {
		for s := range arg.in {
			arg.out <- s
			arg.out <- s
		}
	}

	Print(Numbers(1, 100),
		Grep("3"),
		GrepNot("7"),
		dbl,
		Uniq,
		ReplaceMatch("^(.)$", "x$1"),
		Sort(),
		ReplaceMatch("^(.)", "$1 "),
		dbl,
		DeleteMatch(" .$"),
		UniqWithCount,
		Sort(Num(1)),
		Reverse,
	)

	// Output:
	// 18 3
	// 2 x
	// 2 9
	// 2 8
	// 2 6
	// 2 5
	// 2 4
	// 2 2
	// 2 1
}

func ExampleHash() {
	hash := func(f string, out chan<- string) {
		file, err := os.Open(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		hasher := sha1.New()
		_, err = io.Copy(hasher, file)
		file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		out <- fmt.Sprintf("%x %s", hasher.Sum(nil), f)
	}

	// Some alternative ways of hashing.
	Print(
		Find(FILES, "."),
		Grep("pipe"),
		GrepNot("test"),
		Parallel(4, hash),
		Sort(Text(2)),
	)

	Print(
		System("find", ".", "-type", "f", "-print"),
		Parallel(4, hash),
		Sort(Text(2)),
	)
}
