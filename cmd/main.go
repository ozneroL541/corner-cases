package main

import (
	"encoding/json"
	"fmt"
	"os"

	matrix "matrix"
)

func main() {
	/**
	 * Tests JSON file
	 * TODO: Reference
	 */
	const jsonFile = "test.json"

	// Create JSON if it doesn't exist to ensure reproducible cryptographic inputs
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		fmt.Printf("Generating cryptographic %s file...\n", jsonFile)
		matrix.GenerateInputsFile(jsonFile)
	}
	/** File data */
	fileData, _ := os.ReadFile(jsonFile)
	/** JSON Tests */
	var tests []matrix.TestCase

	json.Unmarshal(fileData, &tests)

	matrix.PrintMatrix(tests)
}
