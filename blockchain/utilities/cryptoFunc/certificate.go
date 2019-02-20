package cryptoFunc

import (
	"errors"
	"math/big"
	"strings"

	"golang.org/x/crypto/bn256"
)

//VerifyBlindCertificate verifies that the blinded certificate is correct
// ie e(b*H(C)*G + b*pubG1CP, b*certificate) == e(b*G, b*pubG2User)
func VerifyBlindCertificate(blindCommitment []byte, blindCertificate []byte, blindPubG1Byte []byte, blindPubG2Byte []byte, blindGenerator []byte) (bool, error) {

	/****** Convert []byte to G1 and G2 bn256 point *****/
	blindPubG1, b := new(bn256.G1).Unmarshal(blindPubG1Byte)
	if b != true {
		return false, errors.New("Error during unmarshal pubG1")
	}
	blindGeneratorPoint, b := new(bn256.G1).Unmarshal(blindGenerator)
	if b != true {
		return false, errors.New("Error during unmarshal generator")
	}
	blindPubG2, b := new(bn256.G2).Unmarshal(blindPubG2Byte)
	if b != true {
		return false, errors.New("Error during unmarshal pubG2")
	}
	blindCertificatePoint, b := new(bn256.G2).Unmarshal(blindCertificate)
	if b != true {
		return false, errors.New("Error during unmarshal certificate")
	}

	//Compute b*H(C)*G1
	blindCommitmentInt := new(big.Int).SetBytes(blindCommitment)
	leftG1 := new(bn256.G1).ScalarBaseMult(blindCommitmentInt)
	//Compute b*H(C) + b*pubG1CP)
	leftG1 = leftG1.Add(leftG1, blindPubG1)

	left := bn256.Pair(leftG1, blindCertificatePoint)
	right := bn256.Pair(blindGeneratorPoint, blindPubG2)

	if strings.Compare(left.String(), right.String()) != 0 {
		return false, nil
	}

	return true, nil
}
