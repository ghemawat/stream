package pipe

import (
	"os"
	"path/filepath"
)

type FindMatch int

const (
	FILES    FindMatch = 1      // Match regular files
	DIRS               = 2      // Match directories
	SYMLINKS           = 4      // Match symbolic links
	ALL                = 0xffff // Match everything
)

// Find copies all input and then produces a sequence of items, one
// per file/directory/symlink found at or under dir that matches mask.
func Find(mask FindMatch, dir string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		err := filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
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
		if err != nil {
			arg.ReportError(err)
		}
	}
}
