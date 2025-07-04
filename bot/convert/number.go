// Package convert provides functions for converting data types
package convert

import (
	"strconv"
	"strings"
)

func StringToInt(s string) (int, error) {
	cleaned := strings.ReplaceAll(strings.ReplaceAll(s, ",", ""), ".", "")
	return strconv.Atoi(cleaned)
}
