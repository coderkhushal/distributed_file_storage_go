package main

import (
	"bytes"
	"io"
	"testing"
)

func TestTransformfunc(t *testing.T) {
	key := "mybestpictures"
	PathKey := CACPathTransformFunc(key)

	expectedFilename := "7037c790557f0d861c53d3bbd1fafe02dc3699e6"
	expectedPathname := "7037c/79055/7f0d8/61c53/d3bbd/1fafe/02dc3/699e6"
	if PathKey.Pathname != expectedPathname {
		t.Errorf("have (%s) want %s ", PathKey.Pathname, expectedPathname)
	}
	if PathKey.Filename != expectedFilename {
		t.Errorf("have (%s) want %s ", PathKey.Filename, expectedFilename)
	}

}
func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CACPathTransformFunc,
		Root:              root,
	}
	key := "somethingspecial"
	s := NewStore(opts)
	data := []byte("some gpeg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}
func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CACPathTransformFunc,
	}
	key := "somethingspecial"
	s := NewStore(opts)
	data := []byte("some gpeg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("have %s want %s ", b, data)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}

}
