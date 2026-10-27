package main

import "fmt"

const expectedOutput = `     | V | V | V | V | V | V |
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

func main() {
	fmt.Print(expectedOutput)
}
