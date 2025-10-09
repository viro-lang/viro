package native

import "strconv"

func formatInt(v int) string {
	return strconv.Itoa(v)
}

func formatUint(v uint64) string {
	return strconv.FormatUint(v, 10)
}
