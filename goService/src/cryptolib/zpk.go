package cryptolib

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
)

//GenerateProof generates a zero knownledge proof for the r value
/*
 * c is the elliptic curve used
 * w random value to be sure that the proof is unique
 * (x, y) is the public generator known by the participant
 * r the secret value we want to proove the knownledge
 *
 * The function output:
 * A random value used to compute and verify the proof
 * The proof
 */
func GenerateProof(c elliptic.Curve, w []byte, x *big.Int, y *big.Int, r []byte) (*ecdsa.PublicKey, []byte) {
	//A = w*(x,y)
	Ax, Ay := c.ScalarMult(x, y, w)
	A := ecdsa.PublicKey{Curve: c, X: Ax, Y: Ay}
	AByte := elliptic.Marshal(c, Ax, Ay)
	h := sha256.New()
	h.Write(AByte)
	//s = H(A)
	s := h.Sum(nil)

	//t = s*r + w [q]
	sInt := new(big.Int).SetBytes(s)
	rInt := new(big.Int).SetBytes(r)
	wInt := new(big.Int).SetBytes(w)

	t := new(big.Int).Mul(sInt, rInt)
	t.Add(t, wInt)
	tByte := t.Bytes()

	return &A, tByte
}

//VerifyProof verifies that a zkp is correct.
/*
 * t the proof
 * A the random value used to compute the proof
 * generator the public generator known by the participant
 * pubSecret the public key multiplied by the random. It has been computed during the proof generation.
 */
func VerifyProof(c elliptic.Curve, t []byte, A *ecdsa.PublicKey, generator *ecdsa.PublicKey, pubSecret *ecdsa.PublicKey) bool {
	leftX, leftY := c.ScalarMult(generator.X, generator.Y, t)

	AByte := elliptic.Marshal(c, A.X, A.Y)
	h := sha256.New()
	h.Write(AByte)
	s := h.Sum(nil)

	rightX, rightY := c.ScalarMult(pubSecret.X, pubSecret.Y, s)
	rightX, rightY = c.Add(rightX, rightY, A.X, A.Y)

	if leftX.Cmp(rightX) != 0 {
		return false
	}
	if leftY.Cmp(rightY) != 0 {
		return false
	}

	return true
}
