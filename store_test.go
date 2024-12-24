package main

import (
	"bytes"
	"testing"
)

func TestTransformfunc(t *testing.T) {
	key := "mybestpictures"
	pathname := CACPathTransformFunc(key)
	expectedPathname := "7037c/79055/7f0d8/61c53/d3bbd/1fafe/02dc3/699e6"
	if pathname != expectedPathname {
		t.Errorf("have (%s) want %s ", pathname, expectedPathname)
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
