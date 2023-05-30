package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func New(root string) FileRunner {
	return FileRunner{
		root:        root,
		include:     map[FilterJoin][]string{},
		exclude:     map[FilterJoin][]string{},
		readContent: false,
	}
}

type FileRunner struct {
	root        string
	include     map[FilterJoin][]string
	exclude     map[FilterJoin][]string
	readContent bool
}

type FilterJoin uint8

const (
	AND FilterJoin = 1 << iota
	OR
)

func (fr FileRunner) Include(joinBy FilterJoin, filters []string) FileRunner {
	for _, f := range filters {
		if strings.HasPrefix(f, "${BASEPATH}") {
			f = strings.Replace(f, "${BASEPATH}", fr.root, 1)
		}
		fr.include[joinBy] = append(fr.include[joinBy], f)
	}
	return fr
}

func (fr FileRunner) Exclude(joinBy FilterJoin, filters []string) FileRunner {
	for _, f := range filters {
		if strings.HasPrefix(f, "${BASEPATH}") {
			f = strings.Replace(f, "${BASEPATH}", fr.root, 1)
		}
		fr.exclude[joinBy] = append(fr.exclude[joinBy], f)
	}
	return fr
}

func (fr FileRunner) ReadContent() FileRunner {
	fr.readContent = true
	return fr
}

func (fr FileRunner) Execute() (files *Files, err error) {
	f := NewFiles(map[string]File{})
	err = filepath.WalkDir(fr.root, fr.filterFunc(&f))
	files = &f
	return
}

func (fr *FileRunner) fulfilsIncludeAND(path string) bool {
	for _, f := range fr.include[AND] {
		if !strings.Contains(path, f) {
			return false
		}
	}
	return true
}
func (fr *FileRunner) fulfilsIncludeOR(path string) bool {
	if len(fr.include[OR]) == 0 {
		return true
	}
	include := false
	for _, f := range fr.include[OR] {
		if strings.Contains(path, f) {
			include = true
			break
		}
	}
	return include
}
func (fr *FileRunner) fulfilsExcludeAND(path string) bool {
	if len(fr.exclude[AND]) == 0 {
		return false
	}
	exclude := true
	for _, f := range fr.exclude[AND] {
		if !strings.Contains(path, f) {
			exclude = false
			break
		}
	}
	return exclude
}

// exclude when at least 1 filter is met
func (fr *FileRunner) fulfilsExcludeOR(path string) bool {
	for _, f := range fr.exclude[OR] {
		if strings.Contains(path, f) {
			return true
		}
	}
	return false
}

func (fr *FileRunner) filterFunc(files *Files) fs.WalkDirFunc {
	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !fr.fulfilsIncludeAND(path) {
			return nil
		}
		if !fr.fulfilsIncludeOR(path) {
			return nil
		}
		if fr.fulfilsExcludeAND(path) {
			return nil
		}
		if fr.fulfilsExcludeOR(path) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		f := File{
			Name:     info.Name(),
			Size:     info.Size(),
			FileMode: info.Mode(),
			IsDir:    info.IsDir(),
			ModTime:  info.ModTime(),
			Content:  []byte{},
		}

		if !fr.readContent || info.IsDir() {
			files.write(path, &f)
			return nil
		}

		c, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f.Content = make([]byte, len(c))
		copy(f.Content, c)
		files.write(path, &f)
		return nil
	}
}

func NewFiles(files map[string]File) Files {
	return Files{files, sync.RWMutex{}}
}

type Files struct {
	m    map[string]File
	lock sync.RWMutex
}

type File struct {
	Name     string      // base name of the file
	Size     int64       // length in bytes for regular files; system-dependent for others
	FileMode fs.FileMode // file mode bits
	IsDir    bool        // modification time
	ModTime  time.Time   // abbreviation for Mode().IsDir()
	Content  []byte
}

func (f *Files) write(key string, value *File) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.m[key] = *value
}

func (f *Files) Content() map[string]File {
	res := make(map[string]File, len(f.m))
	for k, v := range f.m {
		res[k] = v
	}
	return res
}

func (f *Files) Read(key string) (File, bool) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	file, ok := f.m[key]
	return file, ok
}

func (f *Files) CopyContent(key string) ([]byte, bool) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	file, ok := f.m[key]
	if !ok {
		return nil, false
	}
	if file.IsDir {
		return nil, false
	}
	res := make([]byte, 0, len(file.Content))
	res = append(res, file.Content...)
	return res, true
}
