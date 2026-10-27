package matrix

import (
	"crypto/sha512"
	"encoding/hex"

	ed_dsa "ed-dsa"

	"filippo.io/edwards25519"
)

/**
 * DigitalSignature struct
 */
type DigitalSignature struct {
	R []byte `json:"R"`
	S []byte `json:"S"`
}

/**
 * parsePermissiveScalar reads 32 bytes.
 * If non-canonical (S >= L), it returns false for the canonical check.
 * It simulates permissive library behavior by reducing mod L.
 * @param b []byte - 32 bytes representing the scalar S
 * @return (*edwards25519.Scalar, bool) - The parsed scalar and a boolean indicating if it was canonical
 */
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
	// Copy the 32 bytes into the first half of the padded array
	copy(padded[:32], b)
	// Reduce mod L
	s, _ = new(edwards25519.Scalar).SetUniformBytes(padded[:])
	return s, false
}

/**
 * parsePermissivePoint reads 32 bytes. If non-canonical (not a valid point), it returns false for the canonical check.
 * It simulates permissive library behavior by checking for the non-canonical identity point.
 * @param b []byte - 32 bytes representing the point (A or R)
 * @return (*edwards25519.Point, bool) - The parsed point and a boolean indicating if it was canonical
 */
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

/**
 * verifySignature performs the EdDSA signature verification with configurable checks.
 * @param A_bytes []byte - Public key (32 bytes)
 * @param R_bytes []byte - Signature component R (32 bytes)
 * @param S_bytes []byte - Signature component S (32 bytes)
 * @param M []byte - Message to verify
 * @param check_S bool - Whether to check if S is canonical
 * @param check_can bool - Whether to check if A and R are canonical
 * @param check_small bool - Whether to perform small-order check on A
 * @param use_cofactor bool - Whether to use cofactor multiplication in the final comparison
 * @return bool - true if signature is valid, false otherwise
 */
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

// Argument order: check_S, check_can, check_small, use_cofactor

// VER1: S check + canonical check, uncofactored equation
func VER1(verification_key []byte, digital_signature DigitalSignature, message []byte) bool {
	return verifySignature(verification_key, digital_signature.R, digital_signature.S, message, true, true, false, false)
}

// VER2: canonical check only, cofactored equation
func VER2(verification_key []byte, digital_signature DigitalSignature, message []byte) bool {
	return verifySignature(verification_key, digital_signature.R, digital_signature.S, message, false, true, false, true)
}

// VER3: small-order check only, cofactored equation
func VER3(verification_key []byte, digital_signature DigitalSignature, message []byte) bool {
	return verifySignature(verification_key, digital_signature.R, digital_signature.S, message, false, false, true, true)
}

// VER4: canonical + small-order checks, cofactored equation
func VER4(verification_key []byte, digital_signature DigitalSignature, message []byte) bool {
	return verifySignature(verification_key, digital_signature.R, digital_signature.S, message, false, true, true, true)
}

// VER5: S check + small-order check, uncofactored equation
func VER5(verification_key []byte, digital_signature DigitalSignature, message []byte) bool {
	return verifySignature(verification_key, digital_signature.R, digital_signature.S, message, true, false, true, false)
}

// VER6: S check + small-order check, cofactored equation
func VER6(verification_key []byte, digital_signature DigitalSignature, message []byte) bool {
	return verifySignature(verification_key, digital_signature.R, digital_signature.S, message, true, false, true, true)
}

/**
 * Array of verification functions
 */
var VER_FUNC = []func([]byte, DigitalSignature, []byte) bool{VER1, VER2, VER3, VER4, VER5, VER6}
