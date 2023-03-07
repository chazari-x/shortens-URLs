package shortens

import (
	"strconv"
)

func Shortens(n int) string {
	id := strconv.FormatInt(int64(n), 16)

	return id
}
