package matrix

import (
	"encoding/hex"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

/**
 * ComputeMatrix computes the results of the verification functions applied to the test cases.
 * @param tests []TestCase
 * @return [][]bool - A 2D slice where each row corresponds to a test case and each column corresponds to a verification function.
 */
func ComputeMatrix(tests []TestCase) [][]bool {
	var matrix [][]bool = make([][]bool, len(tests))
	for i, t := range tests {
		verification_key, _ := hex.DecodeString(t.A)
		r, _ := hex.DecodeString(t.R)
		s, _ := hex.DecodeString(t.S)
		message, _ := hex.DecodeString(t.Msg)
		digital_signature := DigitalSignature{R: r, S: s}
		matrix[i] = make([]bool, len(VER_FUNC))
		for j := range len(VER_FUNC) {
			matrix[i][j] = VER_FUNC[j](verification_key, digital_signature, message)
		}
	}
	return matrix
}

/**
 * PrintMatrix prints the results of the verification functions applied to the test cases.
 * @param tests []TestCase
 * @param graphical bool - If true, the matrix will be printed graphically. If false, it will be printed on the terminal.
 */
func PrintMatrix(tests []TestCase, graphical bool) {
	// Results of the verification functions applied to the test cases
	results := ComputeMatrix(tests)
	printOnTerminal(results)
	if graphical {
		a := app.New()
		w := a.NewWindow("Verification Matrix")
		w.SetContent(NewBooleanMatrixTable(results))
		w.Resize(fyne.NewSize(imageWidth, imageHeight))
		w.ShowAndRun()
	}
}

/**
 * Graphically print the results of the verification functions applied to the test cases
 * on the terminal.
 * @param results [][]bool - A 2D slice where each row corresponds to a test case and each column corresponds to a verification function.
 */
func printOnTerminal(results [][]bool) {
	fmt.Println("     | V | V | V | V | V | V |")
	fmt.Println("     | E | E | E | E | E | E |")
	fmt.Println("     | R | R | R | R | R | R |")
	fmt.Println("     | 1 | 2 | 3 | 4 | 5 | 6 |")
	fmt.Println("------------------------------")
	for i := range len(results) {
		fmt.Printf("INP%d |", i+1)
		for j := range len(results[i]) {
			if results[i][j] {
				fmt.Print(" ✅|")
			} else {
				fmt.Print("   |")
			}
		}
		fmt.Printf("\n")
	}
}
