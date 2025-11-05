package value

func ClampToRemaining(index, length, requested int) int {
	remaining := length - index
	if requested > remaining {
		return remaining
	}
	if requested < 0 {
		return 0
	}
	return requested
}
