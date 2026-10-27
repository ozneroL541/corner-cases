package matrix

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"

	ed_dsa "ed-dsa"
	utils "utils"

	"filippo.io/edwards25519"
)

/**
 * INP1: Invalid Signature (control case, must be rejected by every verifier)
 * @param tests []TestCase
 * @return []TestCase
 */
func INP1(tests []TestCase) []TestCase {
	sk1, pk1 := ed_dsa.KeyGen()
	msg1 := []byte("Message1")
	r1, s1 := ed_dsa.SigGen(sk1, pk1, msg1)

	// Tamper with S: S' = S + 1 mod L (still canonical, but the equation no longer holds)
	sScalar, _ := new(edwards25519.Scalar).SetCanonicalBytes(s1)
	oneBytes := make([]byte, 32)
	oneBytes[0] = 1
	one, _ := new(edwards25519.Scalar).SetCanonicalBytes(oneBytes)
	sBad := new(edwards25519.Scalar).Add(sScalar, one)

	tests = append(tests, TestCase{"INP1", "Invalid Signature", hex.EncodeToString(msg1), hex.EncodeToString(pk1), hex.EncodeToString(r1), hex.EncodeToString(sBad.Bytes())})
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

	sBig := new(big.Int).SetBytes(utils.ReverseBytes(s2)) // s2 is little-endian
	// L = 2^252 + 27742317777372353535851937790883648493
	L_str := "1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3ed"
	L, _ := new(big.Int).SetString(L_str, 16)
	sBig.Add(sBig, L)

	s2_unreduced := utils.ReverseBytes(utils.PadTo32(sBig.Bytes()))
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
 * INP6: Standard Valid Signature (positive control case)
 * Generates an honest signature and retries until every sanity check passes.
 * @param tests []TestCase
 * @return []TestCase
 */
func INP6(test []TestCase) []TestCase {
	msg6 := []byte("Message6")
	identity := edwards25519.NewIdentityPoint()

	for attempt := 0; attempt < 100; attempt++ {
		sk6, pk6 := ed_dsa.KeyGen()
		r6, s6 := ed_dsa.SigGen(sk6, pk6, msg6)

		// S must be canonical (S < L)
		if _, err := new(edwards25519.Scalar).SetCanonicalBytes(s6); err != nil {
			continue
		}

		// A and R must be canonically encoded and valid points
		if !ed_dsa.IsCanonicalY(pk6) || !ed_dsa.IsCanonicalY(r6) {
			continue
		}
		aPoint, err := new(edwards25519.Point).SetBytes(pk6)
		if err != nil {
			continue
		}
		rPoint, err := new(edwards25519.Point).SetBytes(r6)
		if err != nil {
			continue
		}

		// A must not be a small-order point: [8]A != O
		if new(edwards25519.Point).MultByCofactor(aPoint).Equal(identity) == 1 {
			continue
		}
		// R must not be a small-order point either: [8]R != O
		if new(edwards25519.Point).MultByCofactor(rPoint).Equal(identity) == 1 {
			continue
		}

		// Every verifier must accept the signature
		sig := DigitalSignature{R: r6, S: s6}
		allAccept := true
		for _, ver := range VER_FUNC {
			if !ver(pk6, sig, msg6) {
				allAccept = false
				break
			}
		}
		if !allAccept {
			continue
		}

		test = append(test, TestCase{"INP6", "Standard Valid Signature", hex.EncodeToString(msg6), hex.EncodeToString(pk6), hex.EncodeToString(r6), hex.EncodeToString(s6)})
		return test
	}

	panic("INP6: could not generate a signature passing all checks")
}

/** Array of INP functions */
var INP_FUNC = []func([]TestCase) []TestCase{INP1, INP2, INP3, INP4, INP5, INP6}

/**
 * Generate a JSON file containing the 6 verification functions applied to the 6 test cases.
 */
func GenerateInputsFile(filename string) {
	/** Test Cases for the 6 Verification Functions **/
	var test []TestCase
	for _, inpFunc := range INP_FUNC {
		test = inpFunc(test)
	}
	fileData, _ := json.MarshalIndent(test, "", "  ")
	os.WriteFile(filename, fileData, 0644)
}
