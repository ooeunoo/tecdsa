package encoding

import (
	"encoding/base64"
	"strconv"
)

func Encode(n int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(n)))
}

func Decode(s string) int {
	decoded, _ := base64.StdEncoding.DecodeString(s)
	n, _ := strconv.Atoi(string(decoded))
	return n
}
