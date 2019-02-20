package cryptolib

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"errors"
)

var (
	errMessageOrGeneratorSize = errors.New("Message and generator list must have the same size")
)

//hash a list a messages and return the list of hashed messages
func messageToHash(message [][]byte) [][]byte {
	numMess := len(message)
	hashedMessage := make([][]byte, numMess)
	for i := 0; i < numMess; i++ {
		hashedMessage[i] = make([]byte, 32)
		h256 := sha256.New()
		h256.Write(message[i])
		hashedMessage[i] = h256.Sum(nil)
	}
	return hashedMessage
}

// Commit returns the commitment pedersen of a list of message
func Commit(message [][]byte, secretPub *ecdsa.PublicKey, gen []ecdsa.PublicKey, r []byte) ([]byte, error) {
	//check size of list, must be equal
	if len(message) != len(gen) {
		return nil, errMessageOrGeneratorSize
	}

	//we commit the hash of the message, transform a list of message into a list of hash
	hashedMessage := messageToHash(message)

	//compute the rH part
	resX, resY := secretPub.Curve.ScalarMult(secretPub.X, secretPub.Y, r)

	//compute the m1G1 + ... + mnGn
	for i := 0; i < len(gen); i++ {
		x, y := secretPub.Curve.ScalarMult(gen[i].X, gen[i].Y, hashedMessage[i])
		resX, resY = secretPub.Curve.Add(resX, resY, x, y)
	}

	//compute the []byte form of the result
	res := elliptic.Marshal(secretPub.Curve, resX, resY)
	return res, nil
}

//Verify returns true if the commited value is equal to the computed value.
func Verify(message [][]byte, secretPub *ecdsa.PublicKey, gen []ecdsa.PublicKey, r []byte, c elliptic.Curve, commit []byte) (bool, error) {
	commitComputed, err := Commit(message, secretPub, gen, r)
	if err != nil {
		return false, err
	}
	if bytes.Compare(commit, commitComputed) == 0 {
		return true, nil
	}
	return false, nil
}
