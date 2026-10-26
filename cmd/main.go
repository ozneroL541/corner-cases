package main

import (
	eddsa "ed-dsa"
	"fmt"
	"strings"
)

// InputVector rigorously defines a single test case for the verification matrix.
type InputVector struct {
	M   []byte // Message
	A   []byte // Public Key
	Sig []byte // Signature (R || S)
}

func main() {
	// 1. Define the 6 inputs (INP1 to INP6).
	// Replace these byte slices with the actual test vectors provided in your lab
	// (e.g., malformed signatures, small-order keys, non-canonical scalars).
	inputs := []InputVector{
		{M: []byte("msg1"), A: make([]byte, 32), Sig: make([]byte, 64)}, // INP1
		{M: []byte("msg2"), A: make([]byte, 32), Sig: make([]byte, 64)}, // INP2
		{M: []byte("msg3"), A: make([]byte, 32), Sig: make([]byte, 64)}, // INP3
		{M: []byte("msg4"), A: make([]byte, 32), Sig: make([]byte, 64)}, // INP4
		{M: []byte("msg5"), A: make([]byte, 32), Sig: make([]byte, 64)}, // INP5
		{M: []byte("msg6"), A: make([]byte, 32), Sig: make([]byte, 64)}, // INP6
	}

	// 2. Map the 6 verification functions.
	// Assuming they are defined in the 'eddsa' package as requested previously.
	versions := []func(A, sig, M []byte) bool{
		eddsa.Ver1,
		eddsa.Ver2,
		eddsa.Ver3,
		eddsa.Ver4,
		eddsa.Ver5,
		eddsa.Ver6,
	}

	// 3. Print the header.
	fmt.Println("       VER1 VER2 VER3 VER4 VER5 VER6")

	// 4. Iterate over each input and evaluate it against all 6 verification versions.
	for i, inp := range inputs {
		rowName := fmt.Sprintf("INP%d", i+1)
		var results []string

		for _, verifyFunc := range versions {
			// Evaluate the formal verification function
			isValid := verifyFunc(inp.A, inp.Sig, inp.M)

			// Format the result to match the slide's matrix style
			if isValid {
				results = append(results, "[v]") // Checked box for True
			} else {
				results = append(results, "[ ]") // Empty box for False
			}
		}

		// Print the row rigorously aligned
		fmt.Printf("%s   %s\n", rowName, strings.Join(results, "  "))
	}
}
