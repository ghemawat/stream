package pipe

import (
	"os"
	"path/filepath"
)

// FindFilter is a filter that produces matching nodes under a filesystem
// directory.
type FindFilter struct {
	dir                        string
	seentype                   bool
	files, dirs, symlinks, all bool
	skipdir                    map[string]bool
}

// Find returns a filter that produces matching nodes under a
// filesystem directory.  If no type constraining methods (Files,
// Dirs, Symlinks) are called, all nodes are printed. Otherwise, just
// nodes with a type corresponding to at least one of the called
// methods are printed.
func Find(dir string) *FindFilter {
	return &FindFilter{dir: dir}
}

// Files adjusts f so it matches all regular files.
func (f *FindFilter) Files() *FindFilter {
	f.seentype = true
	f.files = true
	return f
}

// Dirs adjusts f so it matches all directories.
func (f *FindFilter) Dirs() *FindFilter {
	f.seentype = true
	f.dirs = true
	return f
}

// Symlinks adjusts f so it matches all symbolic links.
func (f *FindFilter) Symlinks() *FindFilter {
	f.seentype = true
	f.symlinks = true
	return f
}

// All adjusts f so it matches all types of nodes.
func (f *FindFilter) All() *FindFilter {
	f.seentype = true
	f.all = true
	return f
}

// SkipDir adjusts f so that any node that is one of dirs or a
// descendant of one of the dirs is skipped.
func (f *FindFilter) SkipDir(dirs ...string) *FindFilter {
	if f.skipdir == nil {
		f.skipdir = make(map[string]bool)
	}
	for _, d := range dirs {
		f.skipdir[d] = true
	}
	return f
}

func (f *FindFilter) shouldYield(s os.FileInfo) bool {
	switch {
	case !f.seentype:
		// If no types are specified, match everything
		return true
	case f.all:
		return true
	case f.files && s.Mode().IsRegular():
		return true
	case f.dirs && s.Mode().IsDir():
		return true
	case f.symlinks && s.Mode()&os.ModeSymlink != 0:
		return true
	default:
		return false
	}
}

func (f *FindFilter) RunFilter(arg Arg) error {
	return filepath.Walk(f.dir, func(n string, s os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if f.shouldYield(s) {
			arg.Out <- n
		}
		if f.skipdir != nil && f.skipdir[n] && s.Mode().IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
}
