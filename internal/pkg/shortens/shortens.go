package shortens

import (
	"strconv"
)

func Short(n int) string {
	// получает количество элементов в хранилище
	// преобразует в шестнадцатеричную систему счисления
	// возвращает полученное значение
	id := strconv.FormatInt(int64(n), 16)

	return id
}

func Original(s string) (int, error) {
	id, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
