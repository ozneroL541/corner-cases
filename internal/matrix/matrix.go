package matrix

import (
	"encoding/hex"
	"fmt"
)

/**
 * PrintMatrix prints the results of the verification functions applied to the test cases.
 * @param tests []TestCase
 */
func PrintMatrix(tests []TestCase) {
	fmt.Println("     | V | V | V | V | V | V |")
	fmt.Println("     | E | E | E | E | E | E |")
	fmt.Println("     | R | R | R | R | R | R |")
	fmt.Println("     | 1 | 2 | 3 | 4 | 5 | 6 |")
	fmt.Println("------------------------------")

	for _, t := range tests {
		verification_key, _ := hex.DecodeString(t.A)
		r, _ := hex.DecodeString(t.R)
		s, _ := hex.DecodeString(t.S)
		message, _ := hex.DecodeString(t.Msg)
		digital_signature := DigitalSignature{R: r, S: s}

		fmt.Printf("%s |", t.ID)

		for i := range len(VER_FUNC) {
			if VER_FUNC[i](verification_key, digital_signature, message) {
				fmt.Print(" ✅|")
			} else {
				fmt.Print("   |")
			}
		}
		fmt.Println()
	}
}
