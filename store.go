package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const root = "khushalnetworkss"

func CACPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashstr := hex.EncodeToString(hash[:]) // hash is [20]bytes , hash[:] is []bytes

	blocksize := 5
	slicelen := len(hashstr) / blocksize
	paths := make([]string, slicelen)
	for i := 0; i < slicelen; i++ {
		from, to := i*blocksize, i*blocksize+blocksize
		paths[i] = hashstr[from:to]

	}

	return PathKey{

		Pathname: strings.Join(paths, "/"),
		Filename: hashstr,
	}

}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

type PathTransformFunc func(string) PathKey

type StoreOpts struct {
	// root is the name of root folder in which all files will be stored
	Root              string
	PathTransformFunc PathTransformFunc
}
type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = root
	}
	return &Store{

		StoreOpts: opts,
	}
}
func (s *Store) Has(key string) bool {
	pathkey := CACPathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	if err != nil {
		return false
	}
	return true
}
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}
func (s *Store) Delete(key string) error {
	pathkey := CACPathTransformFunc(key)

	defer func() {
		fmt.Printf("delete %s from disk ", pathkey.Filename)
	}()

	firstpathnameWithRoot := fmt.Sprintf("%s/%s", s.Root, strings.Split(pathkey.Pathname, "/")[0])

	return os.RemoveAll(firstpathnameWithRoot)
}
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err

}
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := CACPathTransformFunc(key)
	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(fullpathWithRoot)
}
func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}
func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	pathKey := s.PathTransformFunc(key)
	pathnameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.Pathname)
	if err := os.MkdirAll(pathnameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}

	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	f, err := os.Create(fullpathWithRoot)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}

	return n, nil
}
