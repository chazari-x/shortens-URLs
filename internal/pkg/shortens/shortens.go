package shortens

import (
	"strconv"
)

func Short(n int) string {
	// получает количество элементов в хранилище
	// преобразует и возвращает полученное значение
	id := strconv.FormatInt(int64(n), 36)

	return id
}

func Original(s string) (int, error) {
	id, err := strconv.ParseInt(s, 36, 64)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
