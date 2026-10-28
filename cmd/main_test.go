package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

const expectedMatrix = `     | V | V | V | V | V | V |
     | E | E | E | E | E | E |
     | R | R | R | R | R | R |
     | 1 | 2 | 3 | 4 | 5 | 6 |
------------------------------
INP1 |   |   |   |   |   |   |
INP2 |   | ✅| ✅| ✅|   |   |
INP3 | ✅| ✅|   |   |   |   |
INP4 |   | ✅| ✅| ✅|   | ✅|
INP5 |   |   | ✅|   | ✅| ✅|
INP6 | ✅| ✅| ✅| ✅| ✅| ✅|
`

func TestMainOutput(t *testing.T) {
	originalStdout := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = write
	t.Cleanup(func() {
		os.Stdout = originalStdout
	})

	if err := exercise(false); err != nil {
		t.Fatalf("exercise failed: %v", err)
	}

	if err := write.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}

	output, err := io.ReadAll(read)
	if err != nil {
		t.Fatalf("read main output: %v", err)
	}
	if err := read.Close(); err != nil {
		t.Fatalf("close stdout pipe reader: %v", err)
	}

	if !bytes.Equal(output, []byte(expectedMatrix)) {
		t.Fatalf("unexpected output:\n%q", output)
	}
}
