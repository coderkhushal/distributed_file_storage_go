package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

func CACPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashstr := hex.EncodeToString(hash[:]) // hash is [20]bytes , hash[:] is []bytes

	blocksize := 5
	slicelen := len(hashstr) / blocksize
	paths := make([]string, slicelen)
	for i := 0; i < slicelen; i++ {
		from, to := i*blocksize, i*blocksize+blocksize
		paths[i] = hashstr[from:to]

	}

	return strings.Join(paths, "/")

}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type PathTransformFunc func(string) string

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}
type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathname := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathname, os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])

	pathAndFilename := pathname + "/" + filename
	f, err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}
	log.Printf("written (%d) bytes to disk  %s ", n, pathAndFilename)
	return nil
}
