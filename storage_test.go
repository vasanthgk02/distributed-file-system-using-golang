package main

import (
	"bytes"
	"testing"
)

func TestCASPathTransform(t *testing.T) {
	key := "mypictures"
	pathKey := CASPathTransform(key)
	expectedOrignalKey := "abac3e6bbd3e468fb226de72127008bab22b8de4"
	expectedPathName := "abac3/e6bbd/3e468/fb226/de721/27008/bab22/b8de4"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedOrignalKey {
		t.Errorf("have %s want %s", pathKey.FileName, expectedOrignalKey)
	}
}

// func TestStore(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransfromFunc: CASPathTransform,
// 	}
// 	s := NewStore(opts)
// 	key := "vasanth kumar"
// 	data := []byte("Some vasanth file string")
// 	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}

// 	r, err := s.Read(key)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	b, _ := io.ReadAll(r)

// 	if string(b) != string(data) {
// 		t.Errorf("want %s have %s", b, data)
// 	}

// }

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransfromFunc: CASPathTransform,
	}
	s := NewStore(opts)
	key := "vasanth kumar"
	data := []byte("Some vasanth file string")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	s.Delete(key)
}

// func TestHas(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransfromFunc: CASPathTransform,
// 	}
// 	s := NewStore(opts)
// 	key := "vasanth kumar.txt"
// 	data := []byte("Some vasanth file string")
// 	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}
// 	check := s.Has(key)
// 	fmt.Println(check)
// }
