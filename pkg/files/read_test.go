package files

import (
	"reflect"
	"sync"
	"testing"
)

type scenario struct {
	title    string
	files    map[string]File
	copyFile string
}

var scenarios = []scenario{
	{
		title: "simple test",
		files: map[string]File{
			"k1/file": {Name: "file", IsDir: false, Content: []byte{1, 2, 3}},
			"k2/dir":  {Name: "dir", IsDir: true, Content: []byte{}},
		},
		copyFile: "k1/file",
	},
}

func TestRead(t *testing.T) {
	for _, s := range scenarios {
		t.Logf("test scenario: %s\n", s.title)
		f := Files{m: s.files, lock: sync.RWMutex{}}

		go read(t, s.files, &f)
		go read(t, s.files, &f)
		go read(t, s.files, &f)

		if _, ok := f.Read("doesNotExist"); ok {
			t.Errorf("Read failure: key %s should not be there but is", "doesNotExist")
			t.Fail()
		}
	}
}

func TestCopy(t *testing.T) {
	for _, s := range scenarios {
		f := Files{m: s.files, lock: sync.RWMutex{}}
		c1, _ := f.CopyContent(s.copyFile)

		if !reflect.DeepEqual(c1, s.files[s.copyFile].Content) {
			t.Errorf("Copy content missmatch: \nwant = \"%+v\"\ngot =  \"%+v\"",
				s.files[s.copyFile].Content, c1)
			t.Fail()
		}
		c1[0] = 100
		if reflect.DeepEqual(c1, s.files[s.copyFile].Content) {
			t.Errorf("Copy by value did not work - values should be different")
			t.Fail()
		}

		const msg = "key \"doesNotExist\" should not be there but is"
		if _, ok := f.Read("doesNotExist"); ok {
			t.Errorf("Read failure: %s", msg)
			t.Fail()
		}
		if _, ok := f.CopyContent("doesNotExist"); ok {
			t.Errorf("Copy failure:  %s", msg)
			t.Fail()
		}
	}
}

func read(t *testing.T, want map[string]File, got *Files) {
	for k, v := range want {
		res, ok := got.Read(k)
		if !ok {
			t.Errorf("Read failure: key %s should be there but is not", k)
			t.Fail()
		}
		if !reflect.DeepEqual(res, v) {
			t.Errorf("Read failure: \nwant = \"%+v\"\ngot =  \"%+v\"", v, res)
			t.Fail()
		}
	}
}
