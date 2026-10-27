package utils

/**
 * ReverseBytes reverses the order of bytes in a slice.
 * @param b The byte slice to reverse.
 * @return A new byte slice with the bytes in reverse order.
 */
func ReverseBytes(b []byte) []byte {
	res := make([]byte, len(b))
	for i := range b {
		res[i] = b[len(b)-1-i]
	}
	return res
}

/**
 * PadTo32 pads a byte slice to 32 bytes by adding leading zeros if necessary.
 * If the input slice is longer than 32 bytes, it returns the last 32 bytes.
 * @param b The byte slice to pad.
 * @return A new byte slice of length 32.
 */
func PadTo32(b []byte) []byte {
	res := make([]byte, 32)
	if len(b) > 32 {
		return b[len(b)-32:]
	}
	copy(res[32-len(b):], b)
	return res
}
