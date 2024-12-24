package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
	_, err := os.Stat(pathkey.FullPath())
	if err != nil {
		return false
	}
	return true
}
func (s *Store) Delete(key string) error {
	pathkey := CACPathTransformFunc(key)

	defer func() {
		fmt.Printf("delete %s from disk ", pathkey.Filename)
	}()

	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FullPath())
	return os.RemoveAll(strings.Split(fullpathWithRoot, "/")[1])
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

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)
	pathnameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.Pathname)
	if err := os.MkdirAll(pathnameWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	f, err := os.Create(fullpathWithRoot)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("written (%d) bytes to disk  %s ", n, fullpathWithRoot)
	return nil
}
