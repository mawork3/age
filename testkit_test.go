// Copyright 2022 The age Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.18
// +build go1.18

package age_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
	"filippo.io/age/armor"
)

//go:generate go test -generate -run ^$

func TestMain(m *testing.M) {
	genFlag := flag.Bool("generate", false, "regenerate test files")
	flag.Parse()
	if *genFlag {
		log.SetFlags(0)
		tests, err := filepath.Glob("testdata/testkit/*")
		if err != nil {
			log.Fatal(err)
		}
		for _, test := range tests {
			os.Remove(test)
		}
		generators, err := filepath.Glob("tests/*.go")
		if err != nil {
			log.Fatal(err)
		}
		for _, generator := range generators {
			vector := strings.TrimSuffix(generator, ".go")
			vector = "testdata/testkit/" + strings.TrimPrefix(vector, "tests/")
			log.Printf("%s -> %s\n", generator, vector)
			out, err := exec.Command("go", "run", generator).Output()
			if err != nil {
				if err, ok := err.(*exec.ExitError); ok {
					log.Fatalf("%s", err.Stderr)
				}
				log.Fatal(err)
			}
			os.WriteFile(vector, out, 0664)
		}
	}

	os.Exit(m.Run())
}

func TestVectors(t *testing.T) {
	tests, err := filepath.Glob("testdata/testkit/*")
	if err != nil {
		log.Fatal(err)
	}
	for _, test := range tests {
		contents, err := os.ReadFile(test)
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimPrefix(test, "testdata/testkit/")
		t.Run(name, func(t *testing.T) {
			testVector(t, contents)
		})
	}
}

func testVector(t *testing.T, test []byte) {
	var (
		expect      string
		payloadHash *[32]byte
		identities  []age.Identity
		armored     bool
	)

	for {
		line, rest, ok := bytes.Cut(test, []byte("\n"))
		if !ok {
			t.Fatal("invalid test file: no payload")
		}
		test = rest
		if len(line) == 0 {
			break
		}
		key, value, _ := strings.Cut(string(line), ": ")
		switch key {
		case "expect":
			switch value {
			case "success":
			case "HMAC failure":
			case "header failure":
			case "armor failure":
			case "payload failure":
			case "no match":
			default:
				t.Fatal("invalid test file: unknown expect value:", value)
			}
			expect = value
		case "payload":
			h, err := hex.DecodeString(value)
			if err != nil {
				t.Fatal(err)
			}
			payloadHash = (*[32]byte)(h)
		case "identity":
			i, err := age.ParseX25519Identity(value)
			if err != nil {
				t.Fatal(err)
			}
			identities = append(identities, i)
		case "passphrase":
			i, err := age.NewScryptIdentity(value)
			if err != nil {
				t.Fatal(err)
			}
			identities = append(identities, i)
		case "armored":
			armored = true
		case "file key":
			// Ignored.
		case "comment":
			t.Log(value)
		default:
			t.Fatal("invalid test file: unknown header key:", key)
		}
	}

	var in io.Reader = bytes.NewReader(test)
	if armored {
		in = armor.NewReader(in)
	}
	r, err := age.Decrypt(in, identities...)
	if err != nil && strings.HasSuffix(err.Error(), "bad header MAC") {
		if expect == "HMAC failure" {
			t.Log(err)
			return
		}
		t.Fatalf("expected %s, got HMAC error", expect)
	} else if e := new(armor.Error); errors.As(err, &e) {
		if expect == "armor failure" {
			t.Log(err)
			return
		}
		t.Fatalf("expected %s, got: %v", expect, err)
	} else if _, ok := err.(*age.NoIdentityMatchError); ok {
		if expect == "no match" {
			t.Log(err)
			return
		}
		t.Fatalf("expected %s, got: %v", expect, err)
	} else if err != nil {
		if expect == "header failure" {
			t.Log(err)
			return
		}
		t.Fatalf("expected %s, got: %v", expect, err)
	} else if expect != "success" && expect != "payload failure" &&
		expect != "armor failure" {
		t.Fatalf("expected %s, got success", expect)
	}
	out, err := io.ReadAll(r)
	if err != nil && expect == "success" {
		t.Fatalf("expected %s, got: %v", expect, err)
	} else if err != nil {
		t.Log(err)
		if expect == "armor failure" {
			if e := new(armor.Error); !errors.As(err, &e) {
				t.Errorf("expected armor.Error, got %T", err)
			}
		}
		if payloadHash != nil && sha256.Sum256(out) != *payloadHash {
			t.Error("partial payload hash mismatch")
		}
		return
	} else if expect != "success" {
		t.Fatalf("expected %s, got success", expect)
	}
	if sha256.Sum256(out) != *payloadHash {
		t.Error("payload hash mismatch")
	}
}
