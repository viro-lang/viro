package value

import "fmt"

func ClampToRemaining(index, length, requested int) (int, error) {
	if requested < 0 {
		return 0, fmt.Errorf("out of bounds: count %d < 0", requested)
	}
	remaining := length - index
	if requested > remaining {
		return remaining, nil
	}
	return requested, nil
}
