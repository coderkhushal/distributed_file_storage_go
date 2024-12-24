package main

import (
	"bytes"
	"testing"
)

func TestTransformfunc(t *testing.T) {
	key := "mybestpictures"
	PathKey := CACPathTransformFunc(key)

	expectedOriginalKey := "7037c790557f0d861c53d3bbd1fafe02dc3699e6"
	expectedPathname := "7037c/79055/7f0d8/61c53/d3bbd/1fafe/02dc3/699e6"
	if PathKey.Pathname != expectedPathname {
		t.Errorf("have (%s) want %s ", PathKey.Pathname, expectedPathname)
	}
	if PathKey.Original != expectedOriginalKey {
		t.Errorf("have (%s) want %s ", PathKey.Original, expectedOriginalKey)
	}

}
func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CACPathTransformFunc,
	}

	s := NewStore(opts)
	data := bytes.NewReader([]byte("some gpeg bytes"))
	if err := s.writeStream("mypicture", data); err != nil {
		t.Error(err)
	}
}
