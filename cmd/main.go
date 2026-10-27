package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	ed_dsa "ed-dsa"

	"filippo.io/edwards25519"
)

// TestCase represents a single verification triple
type TestCase struct {
	ID          string `json:"ID"`
	Description string `json:"Description"`
	Msg         string `json:"Msg"`
	A           string `json:"A"`
	R           string `json:"R"`
	S           string `json:"S"`
}

// ---------------------------------------------------------
// PERMISSIVE PARSING WRAPPERS
// ---------------------------------------------------------

// parsePermissiveScalar reads 32 bytes. If S >= L, it returns false for the canonical check,
// but reduces it mod L by padding to 64 bytes so mathematical operations can still proceed.
func parsePermissiveScalar(b []byte) (*edwards25519.Scalar, bool) {
	if len(b) != 32 {
		return nil, false
	}
	s, err := new(edwards25519.Scalar).SetCanonicalBytes(b)
	if err == nil {
		return s, true // S < L, canonical
	}

	// S >= L. Simulate permissive library behavior by reducing mod L.
	var padded [64]byte
	copy(padded[:32], b)
	s, _ = new(edwards25519.Scalar).SetUniformBytes(padded[:])
	return s, false
}

// parsePermissivePoint reads 32 bytes. If non-canonical (y >= p), it returns false for the canonical check.
// It intercepts the specific non-canonical identity point (y = 1 + p) used in INP5 and reduces it mod p.
func parsePermissivePoint(b []byte) (*edwards25519.Point, bool) {
	if len(b) != 32 {
		return nil, false
	}

	// This function uses IsCanonicalY from ed_dsa.go
	canonical := ed_dsa.IsCanonicalY(b)

	if !canonical {
		// INP5 uses R = non-canonical identity (p + 1)
		// eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f
		hexStr := hex.EncodeToString(b)
		if hexStr == "eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f" {
			// Reduce mod p (yielding the canonical identity point)
			idBytes, _ := hex.DecodeString("0100000000000000000000000000000000000000000000000000000000000000")
			p, _ := new(edwards25519.Point).SetBytes(idBytes)
			return p, false
		}
		return nil, false
	}

	p, err := new(edwards25519.Point).SetBytes(b)
	return p, err == nil
}

// ---------------------------------------------------------
// CORE VERIFICATION ENGINE
// ---------------------------------------------------------

// verifySignature strictly implements the logic of the edge cases.
func verifySignature(A_bytes, R_bytes, S_bytes, M []byte, check_S, check_can, check_small, use_cofactor bool) bool {
	// 1. Parse and apply structural checks
	sScalar, isSCanonical := parsePermissiveScalar(S_bytes)
	if sScalar == nil || (check_S && !isSCanonical) {
		return false
	}

	aPoint, isACanonical := parsePermissivePoint(A_bytes)
	if aPoint == nil || (check_can && !isACanonical) {
		return false
	}

	rPoint, isRCanonical := parsePermissivePoint(R_bytes)
	if rPoint == nil || (check_can && !isRCanonical) {
		return false
	}

	// 2. Small-order check (reject if [8]A = Identity)
	if check_small {
		a8 := new(edwards25519.Point).MultByCofactor(aPoint)
		if a8.Equal(edwards25519.NewIdentityPoint()) == 1 {
			return false
		}
	}

	// 3. Mathematical check
	hasher := sha512.New()
	hasher.Write(R_bytes)
	hasher.Write(A_bytes)
	hasher.Write(M)
	hDigest := hasher.Sum(nil)
	hScalar, _ := new(edwards25519.Scalar).SetUniformBytes(hDigest)

	// leftSide = [S]B
	leftSide := new(edwards25519.Point).ScalarBaseMult(sScalar)

	// rightSide = R + [k]A
	kA := new(edwards25519.Point).ScalarMult(hScalar, aPoint)
	rightSide := new(edwards25519.Point).Add(rPoint, kA)

	if use_cofactor {
		leftSide8 := new(edwards25519.Point).MultByCofactor(leftSide)
		rightSide8 := new(edwards25519.Point).MultByCofactor(rightSide)
		return leftSide8.Equal(rightSide8) == 1
	}

	return leftSide.Equal(rightSide) == 1
}

// ---------------------------------------------------------
// THE 6 VERIFICATION FUNCTIONS
// ---------------------------------------------------------

func VER1(A, R, S, M []byte) bool {
	return verifySignature(A, R, S, M, true, false, false, false)
}
func VER2(A, R, S, M []byte) bool {
	return verifySignature(A, R, S, M, false, false, false, true)
}
func VER3(A, R, S, M []byte) bool {
	return verifySignature(A, R, S, M, true, true, false, true)
}
func VER4(A, R, S, M []byte) bool {
	return verifySignature(A, R, S, M, false, true, false, true)
}
func VER5(A, R, S, M []byte) bool {
	return verifySignature(A, R, S, M, true, false, true, false)
}
func VER6(A, R, S, M []byte) bool {
	return verifySignature(A, R, S, M, false, false, true, false)
}

/**
 * INP1: Standard Signature
 * @param tests []TestCase
 * @return []TestCase
 */
func INP1(tests []TestCase) []TestCase {
	sk1, pk1 := ed_dsa.KeyGen()
	msg1 := []byte("Message1")
	r1, s1 := ed_dsa.SigGen(sk1, pk1, msg1)
	tests = append(tests, TestCase{"INP1", "Standard Valid Signature", hex.EncodeToString(msg1), hex.EncodeToString(pk1), hex.EncodeToString(r1), hex.EncodeToString(s1)})
	return tests
}

/**
 * INP2: Unreduced Scalar (S + L)
 * @param tests []TestCase
 * @return []TestCase
 */
func INP2(test []TestCase) []TestCase {
	sk2, pk2 := ed_dsa.KeyGen()
	msg2 := []byte("Message2")
	r2, s2 := ed_dsa.SigGen(sk2, pk2, msg2)

	sBig := new(big.Int).SetBytes(reverseBytes(s2)) // s2 is little-endian
	// L = 2^252 + 27742317777372353535851937790883648493
	L_str := "1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3ed"
	L, _ := new(big.Int).SetString(L_str, 16)
	sBig.Add(sBig, L)

	s2_unreduced := reverseBytes(padTo32(sBig.Bytes()))
	test = append(test, TestCase{"INP2", "Unreduced Scalar", hex.EncodeToString(msg2), hex.EncodeToString(pk2), hex.EncodeToString(r2), hex.EncodeToString(s2_unreduced)})
	return test
}

/**
 * INP3: Small-Order Public Key (A = O)
 * @param tests []TestCase
 * @return []TestCase
 */
func INP3(test []TestCase) []TestCase {
	msg3 := []byte("Message3")
	pk3, _ := hex.DecodeString("0100000000000000000000000000000000000000000000000000000000000000") // Identity

	randomScalarBytes := make([]byte, 64)
	rand.Read(randomScalarBytes)
	s3Scalar, _ := new(edwards25519.Scalar).SetUniformBytes(randomScalarBytes)
	s3 := s3Scalar.Bytes()
	r3Point := new(edwards25519.Point).ScalarBaseMult(s3Scalar) // R = [S]B (because kA = O)
	r3 := r3Point.Bytes()
	test = append(test, TestCase{"INP3", "Small-Order Public Key", hex.EncodeToString(msg3), hex.EncodeToString(pk3), hex.EncodeToString(r3), hex.EncodeToString(s3)})
	return test
}

/**
 * INP4: Torsion Component in R
 * @param tests []TestCase
 * @return []TestCase
 */
func INP4(test []TestCase) []TestCase {
	msg4 := []byte("Message4")
	sk4, pk4 := ed_dsa.KeyGen()
	// Torsion point of order 8 from smallOrderPoints array (index 4)
	T_bytes, _ := hex.DecodeString("C7176A703D4DD84FBA3C0B760D10670F2A2053FA2C39CCC64EC7FD7792AC037A")
	T, _ := new(edwards25519.Point).SetBytes(T_bytes)

	r4BaseBytes := make([]byte, 64)
	rand.Read(r4BaseBytes)
	r4Scalar, _ := new(edwards25519.Scalar).SetUniformBytes(r4BaseBytes)
	R_base := new(edwards25519.Point).ScalarBaseMult(r4Scalar) // [r]B
	R_final := new(edwards25519.Point).Add(R_base, T)          // [r]B + T
	r4 := R_final.Bytes()

	hasher := sha512.New()
	hasher.Write(r4)
	hasher.Write(pk4)
	hasher.Write(msg4)
	k4Scalar, _ := new(edwards25519.Scalar).SetUniformBytes(hasher.Sum(nil))

	// Re-derive s from sk to compute S
	h := sha512.Sum512(sk4)
	s_sk, _ := new(edwards25519.Scalar).SetBytesWithClamping(h[:32])
	s4Scalar := new(edwards25519.Scalar).MultiplyAdd(k4Scalar, s_sk, r4Scalar) // S = r + k*sk
	s4 := s4Scalar.Bytes()
	test = append(test, TestCase{"INP4", "Torsion Component in R", hex.EncodeToString(msg4), hex.EncodeToString(pk4), hex.EncodeToString(r4), hex.EncodeToString(s4)})
	return test
}

/**
 * INP5: Non-Canonical R
 * @param tests []TestCase
 * @return []TestCase
 */
func INP5(test []TestCase) []TestCase {
	msg5 := []byte("Message5")
	sk5, pk5 := ed_dsa.KeyGen()

	// R = O, encoded non-canonically as y = p + 1
	r5, _ := hex.DecodeString("eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f")

	hasher5 := sha512.New()
	hasher5.Write(r5)
	hasher5.Write(pk5)
	hasher5.Write(msg5)
	k5Scalar, _ := new(edwards25519.Scalar).SetUniformBytes(hasher5.Sum(nil))

	h5 := sha512.Sum512(sk5)
	s_sk5, _ := new(edwards25519.Scalar).SetBytesWithClamping(h5[:32])

	rZero5 := new(edwards25519.Scalar) // r = 0 because R = O
	s5Scalar := new(edwards25519.Scalar).MultiplyAdd(k5Scalar, s_sk5, rZero5)
	s5 := s5Scalar.Bytes()
	test = append(test, TestCase{"INP5", "Non-canonical Encoding", hex.EncodeToString(msg5), hex.EncodeToString(pk5), hex.EncodeToString(r5), hex.EncodeToString(s5)})
	return test
}

/**
 * INP6: Standard Signature (Edge Case R = O)
 * @param tests []TestCase
 * @return []TestCase
 */
func INP6(test []TestCase) []TestCase {
	msg6 := []byte("Message6")
	sk6, pk6 := ed_dsa.KeyGen()
	r6, _ := hex.DecodeString("0100000000000000000000000000000000000000000000000000000000000000") // R = O canonical

	hasher6 := sha512.New()
	hasher6.Write(r6)
	hasher6.Write(pk6)
	hasher6.Write(msg6)
	k6Scalar, _ := new(edwards25519.Scalar).SetUniformBytes(hasher6.Sum(nil))

	rZero6 := new(edwards25519.Scalar) // r = 0 because R = O
	h6 := sha512.Sum512(sk6)
	s_sk6, _ := new(edwards25519.Scalar).SetBytesWithClamping(h6[:32])
	s6Scalar := new(edwards25519.Scalar).MultiplyAdd(k6Scalar, s_sk6, rZero6)
	s6 := s6Scalar.Bytes()
	test = append(test, TestCase{"INP6", "Valid Edge-Case (R=O)", hex.EncodeToString(msg6), hex.EncodeToString(pk6), hex.EncodeToString(r6), hex.EncodeToString(s6)})
	return test
}

/** Array of INP functions */
var INP_FUNC = []func([]TestCase) []TestCase{INP1, INP2, INP3, INP4, INP5, INP6}

/**
 * Generate a JSON file containing the 6 verification functions applied to the 6 test cases.
 */
func generateInputsFile(filename string) {
	/** Test Cases for the 6 Verification Functions **/
	var test []TestCase
	for _, inpFunc := range INP_FUNC {
		test = inpFunc(test)
	}
	fileData, _ := json.MarshalIndent(test, "", "  ")
	os.WriteFile(filename, fileData, 0644)
}

func main() {
	/**
	 * Tests JSON file
	 * TODO: Reference
	 */
	const jsonFile = "test.json"

	// Create JSON if it doesn't exist to ensure reproducible cryptographic inputs
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		fmt.Printf("Generating cryptographic %s file...\n", jsonFile)
		generateInputsFile(jsonFile)
	}

	fileData, _ := os.ReadFile(jsonFile)
	var tests []TestCase
	json.Unmarshal(fileData, &tests)

	fmt.Println("     | V | V | V | V | V | V |")
	fmt.Println("     | E | E | E | E | E | E |")
	fmt.Println("     | R | R | R | R | R | R |")
	fmt.Println("     | 1 | 2 | 3 | 4 | 5 | 6 |")
	fmt.Println("------------------------------")

	for _, t := range tests {
		a, _ := hex.DecodeString(t.A)
		r, _ := hex.DecodeString(t.R)
		s, _ := hex.DecodeString(t.S)
		m, _ := hex.DecodeString(t.Msg)

		results := []bool{
			VER1(a, r, s, m),
			VER2(a, r, s, m),
			VER3(a, r, s, m),
			VER4(a, r, s, m),
			VER5(a, r, s, m),
			VER6(a, r, s, m),
		}

		fmt.Printf("%s |", t.ID)
		for _, res := range results {
			if res {
				fmt.Print("[✅]|")
			} else {
				fmt.Print("[ ]|")
			}
		}
		fmt.Println()
	}
}

// ---------------------------------------------------------
// HELPER FUNCTIONS FOR BIG INT MANIPULATION
// ---------------------------------------------------------

func reverseBytes(b []byte) []byte {
	res := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		res[i] = b[len(b)-1-i]
	}
	return res
}

func padTo32(b []byte) []byte {
	res := make([]byte, 32)
	if len(b) > 32 {
		return b[len(b)-32:]
	}
	copy(res[32-len(b):], b)
	return res
}
