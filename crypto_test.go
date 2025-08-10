package main

import (
	"bytes"
	"testing"
)

func TestCopyCrypto(t *testing.T) {
	payload := "Vasanth Kumar"
	key := newEncryptionKey()
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)

	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	// t.Log(dst.Bytes())

	out := new(bytes.Buffer)
	if _, err := copyDecrypt(key, dst, out); err != nil {
		t.Error(err)
	}

	// t.Log(out.String())

	if out.String() != payload {
		t.Errorf("want %s actual %s", payload, out.String())
	}

}
