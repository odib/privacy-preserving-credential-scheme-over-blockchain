package cryptolib

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"strings"

	"golang.org/x/crypto/bn256"
)

//GeneratePairingKey generate a private key and two public keys. (Because Pairing goes from G1 x G2 -> G3)
func GeneratePairingKey() (priv []byte, g1Pub []byte, g2Pub []byte, err error) {
	privInt, err := rand.Int(rand.Reader, bn256.Order)
	if err != nil {
		return nil, nil, nil, err
	}

	g1PubInt := new(bn256.G1).ScalarBaseMult(privInt)
	g2PubInt := new(bn256.G2).ScalarBaseMult(privInt)

	priv = privInt.Bytes()
	g1Pub = g1PubInt.Marshal()
	g2Pub = g2PubInt.Marshal()
	return
}

//GenerateCertificate generates a certificate for the commitment.
//priv is the private key of the certificate provider and pubG2Byte the public key for the owner of the commitment
func GenerateCertificate(commitment []byte, priv []byte, pubG2Byte []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(commitment)
	hash := h.Sum(nil)

	pubG2, err := new(bn256.G2).Unmarshal(pubG2Byte)
	if err != true {
		return nil, errors.New("Cannot Unmarshale pubG2Byte")
	}

	//Compute (H(C)+priv)^{-1}
	hashInt := new(big.Int).SetBytes(hash)
	privInt := new(big.Int).SetBytes(priv)
	certInt := new(big.Int).Add(hashInt, privInt)
	certInt = certInt.ModInverse(certInt, bn256.Order)

	//(H(C)+priv)^{-1}*pubG2
	cert := new(bn256.G2).ScalarMult(pubG2, certInt)
	certByte := cert.Marshal()
	return certByte, nil
}

//VerifyCertificate verifies that the certificate is well formed
//commitment is the commitment used to construct the certificate
//pubG1Byte is the public key of the certificate provider
//pubG2Byte is the public key of the owner of the certificate
func VerifyCertificate(commitment []byte, certificate []byte, pubG1Byte []byte, pubG2Byte []byte) (bool, error) {
	h := sha256.New()
	h.Write(commitment)
	hash := h.Sum(nil)

	pubG2, err := new(bn256.G2).Unmarshal(pubG2Byte)
	if err != true {
		return false, errors.New("Cannot Unmarshal pubG2Byte")
	}

	pubG1, err := new(bn256.G1).Unmarshal(pubG1Byte)
	if err != true {
		return false, errors.New("Cannot Unmarshal pubG1Byte")
	}

	certG2, err := new(bn256.G2).Unmarshal(certificate)
	if err != true {
		return false, errors.New("Cannot Unmarshal certificate")
	}
	// We want to verify that e(H(C)*G + pubG1CP, certificate) == e(G, pubG2User)
	//We compute the G1 term on the left equality ie H(C)*G + pubG1CP
	leftG1 := new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(hash))

	leftG1 = leftG1.Add(leftG1, pubG1)
	//left term ie e(H(C)*G + pubG1CP, certificate)
	left := bn256.Pair(leftG1, certG2)

	//We compute the right term
	//We get back the generator G by doing a scalarBaseMult of 1
	rightG1 := new(bn256.G1).ScalarBaseMult(new(big.Int).SetInt64(1))
	right := bn256.Pair(rightG1, pubG2)

	if strings.Compare(left.String(), right.String()) != 0 {
		return false, nil
	}

	return true, nil
}

//BlindCertificate generates a random and return the input blinded + the generator blinded and the random factor
/* commitment is the commitment used to generate the certificate
 * certificate is the certificate
 * pubG1Byte is the public key of the certificate provider
 * pubG2Byte is the public key of the user
 * privUserByte is the private key of the user
 *
 * The function output (in order of output):
 * The blinded commitment
 * The blinded certificate
 * The blinded public key of the certificate provider
 * The blinded public key of the user
 * The blinded private key of the user
 * The blinded public generator used in the elliptic curve
 * The random used to blind
 */
func BlindCertificate(commitment []byte, certificate []byte, pubG1Byte []byte, pubG2Byte []byte, privUserByte []byte) ([]byte, []byte, []byte, []byte, []byte, []byte, []byte) {
	random, _ := rand.Int(rand.Reader, bn256.Order)

	h := sha256.New()
	h.Write(commitment)
	hashCommitment := h.Sum(nil)

	commitmentInt := new(big.Int).SetBytes(hashCommitment)
	privUserInt := new(big.Int).SetBytes(privUserByte)

	commitmentInt = commitmentInt.Mul(commitmentInt, random)

	certificatePoint, _ := new(bn256.G2).Unmarshal(certificate)
	certificatePoint = certificatePoint.ScalarMult(certificatePoint, random)
	certificateByte := certificatePoint.Marshal()

	pubG1Point, _ := new(bn256.G1).Unmarshal(pubG1Byte)
	pubG1Point = pubG1Point.ScalarMult(pubG1Point, random)
	g1Byte := pubG1Point.Marshal()

	pubG2Point, _ := new(bn256.G2).Unmarshal(pubG2Byte)
	pubG2Point = pubG2Point.ScalarMult(pubG2Point, random)
	g2Byte := pubG2Point.Marshal()

	privUserInt = privUserInt.Mul(privUserInt, random)

	generator := new(bn256.G1).ScalarBaseMult(random)
	generatorByte := generator.Marshal()

	return commitmentInt.Bytes(), certificateByte, g1Byte, g2Byte, privUserInt.Bytes(), generatorByte, random.Bytes()
}

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
