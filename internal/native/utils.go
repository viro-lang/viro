package native

import "strconv"

func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func formatUint(v uint64) string {
	return strconv.FormatUint(v, 10)
}
