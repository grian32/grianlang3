package util

// IntegerInRange reports whether a sign and unsigned magnitude fit the target
// integer type. bitSize must be between 1 and 64. Keeping the sign separate lets
// callers check signed minima and values above MaxInt64 without overflowing.
func IntegerInRange(magnitude uint64, negative bool, bitSize uint8, signed bool) bool {
	if bitSize == 0 || bitSize > 64 {
		return false
	}
	if signed {
		limit := uint64(1) << (bitSize - 1)
		if negative {
			return magnitude <= limit
		}
		return magnitude < limit
	}
	if negative && magnitude != 0 {
		return false
	}
	return bitSize == 64 || magnitude < uint64(1)<<bitSize
}
