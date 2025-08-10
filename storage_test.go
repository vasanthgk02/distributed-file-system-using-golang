package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestCASPathTransform(t *testing.T) {
	key := "mypictures"
	pathKey := CASPathTransform(key)
	expectedFileName := "abac3e6bbd3e468fb226de72127008bab22b8de4"
	expectedPathName := "abac3/e6bbd/3e468/fb226/de721/27008/bab22/b8de4"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedFileName {
		t.Errorf("have %s want %s", pathKey.FileName, expectedFileName)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer tearDown(t, s)

	for i := range 50 {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("Some vasanth file string")

		if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)

		if string(b) != string(data) {
			t.Errorf("want %s have %s", b, data)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("expected NOT to have key %s", key)
		}

	}

}

// func TestDelete(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransfromFunc: CASPathTransform,
// 	}
// 	s := NewStore(opts)
// 	key := "vasanth kumar"
// 	data := []byte("Some vasanth file string")
// 	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}
// 	s.Delete(key)
// }

// func TestHas(t *testing.T) {
// 	s := newStore()
// 	key := "vasanth kumar.txt"
// 	data := []byte("Some vasanth file string")
// 	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}
// 	check := s.Has(key)
// 	log.Println(check)
// }

func newStore() *Store {
	opts := StoreOpts{
		PathTransfromFunc: CASPathTransform,
	}
	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error("error while tearing down....")
	}
}
