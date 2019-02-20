package util

import "encoding/hex"

//HexToByte converts the string containing an hexa decimal value into a byte representation
func HexToByte(s string) ([]byte, error) {
	var err error
	l := hex.DecodedLen(len(s))
	res := make([]byte, l+(len(s)%2))
	if (len(s) % 2) == 1 {
		_, err = hex.Decode(res, []byte("0"+s))
	} else {
		_, err = hex.Decode(res, []byte(s))
	}
	return res, err
}
