package main

import (
	"encoding/json"
	"matrix"
	"os"
)

func main() {
	const gui bool = true
	/**
	 * Tests JSON file
	 */
	exercise(gui)
}

func exercise(gui bool) {
	const jsonFile = "test.json"

	// Create JSON if it doesn't exist to ensure reproducible cryptographic inputs
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		matrix.GenerateInputsFile(jsonFile)
	}
	/** File data */
	fileData, _ := os.ReadFile(jsonFile)
	/** JSON Tests */
	var tests []matrix.TestCase

	json.Unmarshal(fileData, &tests)

	matrix.PrintMatrix(tests, gui)
}
