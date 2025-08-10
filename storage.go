package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Default ROOT path
const DEFAULT_ROOT_FOLDER_NAME string = "vasanthnetwork"

type PathKey struct {
	PathName string
	FileName string
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

func (p PathKey) FirstPathName() string {
	rootPath := strings.Split(p.PathName, "/")
	if len(rootPath) == 0 {
		return ""
	}
	return rootPath[0]
}

type PathTransfromFunc func(string) PathKey

func CASPathTransform(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := range sliceLen {
		from := i * blockSize
		to := from + blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}

}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

type StoreOpts struct {
	Root              string // Root is the folder name of the root, containings all the folder/files of the system.
	PathTransfromFunc PathTransfromFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransfromFunc == nil {
		opts.PathTransfromFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = DEFAULT_ROOT_FOLDER_NAME
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransfromFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk\n", pathKey.FileName)
	}()
	firstPathnameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())
	return os.RemoveAll(firstPathnameWithRoot)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransfromFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) writeDecrypt(encKey []byte, key string, r io.Reader) (int, error) {
	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)
	return n, err
}

func (s *Store) openFileForWriting(key string) (*os.File, error) {
	pathKey := s.PathTransfromFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathKeyWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Create(fullPathWithRoot)
}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransfromFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	file, err := os.Open(pathKeyWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fileinfo, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fileinfo.Size(), file, err
}
