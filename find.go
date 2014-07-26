package pipe

import (
	"os"
	"path/filepath"
)

// FindMatch is a bit mask that selects the types of filesystem nodes
// (files, directories, etc.) that should be yielded by Find.
type FindMatch int

// Values that can be or-ed together and passed to Find to match different
// types of filesystem nodes.
const (
	FILES    FindMatch = 1
	DIRS               = 2
	SYMLINKS           = 4
	ALL                = 0xffff
)

// Find produces a sequence of items, one per file/directory/symlink
// found at or under dir that matches mask.
func Find(mask FindMatch, dir string) Filter {
	return FilterFunc(func(arg Arg) error {
		return filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			if mask&ALL == ALL ||
				mask&FILES != 0 && s.Mode().IsRegular() ||
				mask&DIRS != 0 && s.Mode().IsDir() ||
				mask&SYMLINKS != 0 && s.Mode()&os.ModeSymlink != 0 {
				arg.Out <- f
			}
			return nil
		})
	})
}
