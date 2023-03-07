package shortens

import "strconv"

func Shortens(n int) string {
	pref := 0
	id := ""
	for {
		if n > 90 {
			n -= 90
			pref += 1
		} else {
			break
		}
	}

	for i := 1; i <= 6; i++ {
		switch n {
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9:
			id = strconv.Itoa(n) + id
			n = 0
		case 10:
			id = "a" + id
			n = 0
		case 11:
			id = "b" + id
			n = 0
		case 12:
			id = "c" + id
			n = 0
		case 13:
			id = "d" + id
			n = 0
		case 14:
			id = "e" + id
			n = 0
		case 15:
			id = "f" + id
			n = 0
		default:
			id = "f" + id
			n -= 15
		}
	}

	return strconv.Itoa(pref) + "/" + id
}
