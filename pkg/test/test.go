package eddsa

import (
	"crypto/ed25519"
	"math/big"
)

// ----------------------------------------------------------------------------
// VER1: Standard RFC 8032 Verification
// Formally verifies the cofactorless equation: S * B = R + c * A
// Relies entirely on the underlying Go crypto/ed25519 implementation.
// ----------------------------------------------------------------------------
func Ver1(A, sig, M []byte) bool {
	if len(sig) != 64 || len(A) != 32 {
		return false
	}
	return ed25519.Verify(A, M, sig)
}

// ----------------------------------------------------------------------------
// VER2: Standard Verification + Strict Canonical S Check
// Formally enforces S < L. While modern Go's ed25519 does this internally,
// this explicitly implements the rejection of non-canonical S via big.Int.
// ----------------------------------------------------------------------------
func Ver2(A, sig, M []byte) bool {
	if len(sig) != 64 || len(A) != 32 {
		return false
	}
	S := sig[32:]
	
	sInt := new(big.Int).SetBytes(reverseBytes(S))
	if sInt.Cmp(L) >= 0 || sInt.Sign() < 0 {
		return false
	}

	return ed25519.Verify(A, M, sig)
}

// ----------------------------------------------------------------------------
// VER3: Standard Verification + Reject Small Order Public Keys
// Formally enforces A ∉ {Small Order Points}. Prevents attacks where the
// public key resides in the prime-order subgroup torsion.
// ----------------------------------------------------------------------------
func Ver3(A, sig, M []byte) bool {
	if len(sig) != 64 || len(A) != 32 {
		return false
	}

	for _, smallPoint := range smallOrderPoints {
		if bytesEqual(A, smallPoint) {
			return false
		}
	}

	return ed25519.Verify(A, M, sig)
}

// ----------------------------------------------------------------------------
// VER4: Standard Verification + Canonical Y-Coordinate Check (A and R)
// Formally enforces that the y-coordinates of A and R are strictly reduced
// modulo p = 2^255 - 19.
// ----------------------------------------------------------------------------
func Ver4(A, sig, M []byte) bool {
	if len(sig) != 64 || len(A) != 32 {
		return false
	}
	R := sig[:32]

	if !IsCanonicalY(A) || !IsCanonicalY(R) {
		return false
	}

	return ed25519.Verify(A, M, sig)
}

// ----------------------------------------------------------------------------
// VER5: Reject Small Order A + Reject Non-Canonical A and R
// Combines topological point validation (canonicity) with subgroup checks.
// ----------------------------------------------------------------------------
func Ver5(A, sig, M []byte) bool {
	if len(sig) != 64 || len(A) != 32 {
		return false
	}
	R := sig[:32]

	for _, smallPoint := range smallOrderPoints {
		if bytesEqual(A, smallPoint) {
			return false
		}
	}

	if !IsCanonicalY(A) || !IsCanonicalY(R) {
		return false
	}

	return ed25519.Verify(A, M, sig)
}

// ----------------------------------------------------------------------------
// VER6: Strict Full Verification (Matches your SignVrfy logic)
// Formally enforces: S < L ∧ A ∉ {Small Order} ∧ Canonical(A) ∧ Canonical(R).
// ----------------------------------------------------------------------------
func Ver6(A, sig, M []byte) bool {
	if len(sig) != 64 || len(A) != 32 {
		return false
	}
	R := sig[:32]
	S := sig[32:]

	// 1. Check S < L
	sInt := new(big.Int).SetBytes(reverseBytes(S))
	if sInt.Cmp(L) >= 0 || sInt.Sign() < 0 {
		return false
	}

	// 2. Check A is not small order
	for _, smallPoint := range smallOrderPoints {
		if bytesEqual(A, smallPoint) {
			return false
		}
	}

	// 3. Check A and R are canonical
	if !IsCanonicalY(A) || !IsCanonicalY(R) {
		return false
	}

	// 4. Base verification
	return ed25519.Verify(A, M, sig)
}