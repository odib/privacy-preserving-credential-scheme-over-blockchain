package cryptolib

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

/*
 * Test vector for the commitment. We test first single message commitment. Then multiple message commitment
 * !!! BE CAREFULL !!! During the generation of the test vector, when an elliptic point begin with one (or two) 0x00, they has not been written.
 * If you do not add the 0x00 when the length of the point is not 64, hash will differ and commitment too.
 * In this file, the function parseLine do the job
 */

func parseLine(l string) string {
	var lineParsed string
	a := 64 - len(l)
	if a != 0 {
		lineParsed = "0" + l
		for k := 1; k < a; k++ {
			lineParsed = "0" + lineParsed
		}
	} else {
		lineParsed = l
	}
	return lineParsed
}

func TestSimpleCommitment(t *testing.T) {
	f, err := os.Open("./testFile/SimpleCommitment.vector")

	if err != nil {
		t.Fatal(err)
	}
	buf := bufio.NewReader(f)
	lineNo := 1

	var message []byte
	var keyX []byte
	var keyY []byte
	var genX []byte
	var genY []byte
	var random []byte
	var commitment []byte
	var lineParsed string

	for {

		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			t.Fatalf("error reading from input: %s", err)
		}

		lineNo++

		if !strings.HasSuffix(line, "\n") {
			t.Fatalf("bad line ending (expected \\n) on line %d", lineNo)
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		line = line[:len(line)-1]

		switch {
		case strings.HasPrefix(line, "Message = "):
			message, err = hex.DecodeString(line[10:])
		case strings.HasPrefix(line, "X = "):
			lineParsed = parseLine(line[4:])
			keyX, err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "Y = "):
			lineParsed = parseLine(line[4:])
			keyY, err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "GenX = "):
			lineParsed = parseLine(line[7:])
			genX, err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "GenY = "):
			lineParsed = parseLine(line[7:])
			genY, err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "Random = "):
			random, err = hex.DecodeString(line[9:])
		case strings.HasPrefix(line, "Commitment = "):
			commitment, err = hex.DecodeString(line[13:])

			var key ecdsa.PublicKey
			var gen ecdsa.PublicKey
			key.X = new(big.Int).SetBytes(keyX)
			if err != nil {
				t.Fatalf("Error during commitment parsing at line: %d: %v", lineNo, err)
			}
			key.Y = new(big.Int).SetBytes(keyY)
			if err != nil {
				t.Fatalf("Error during commitment parsing at line: %d: %v", lineNo, err)
			}
			gen.X = new(big.Int).SetBytes(genX)
			gen.Y = new(big.Int).SetBytes(genY)

			key.Curve = crypto.S256()
			gen.Curve = crypto.S256()

			computedCommitment, err := Commit([][]byte{message}, &key, []ecdsa.PublicKey{gen}, random)
			if err != nil {
				t.Fatalf("Error during commitment computation at line: %d: %v", lineNo, err)
			}

			if bytes.Compare(commitment, computedCommitment) != 0 {
				t.Fatalf("Bad commitment at line: %d\n Got: %x\n Want:%x", lineNo, computedCommitment, commitment)
			}
		default:
		}
	}
}

func TestMultipleCommitment(t *testing.T) {
	f, err := os.Open("./testFile/MultipleCommitment.vector")

	if err != nil {
		t.Fatal(err)
	}
	buf := bufio.NewReader(f)
	lineNo := 1

	var message [][]byte
	var keyX []byte
	var keyY []byte
	var genX [][]byte
	var genY [][]byte
	var random []byte
	var commitment []byte
	var lineParsed string
	var indice int
	indiceMess := 0
	indiceGen := 0

	for {

		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			t.Fatalf("error reading from input: %s", err)
		}

		lineNo++

		if !strings.HasSuffix(line, "\n") {
			t.Fatalf("bad line ending (expected \\n) on line %d", lineNo)
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if line[0] == '[' {
			tmp := line[10:]
			indice, err = strconv.Atoi(tmp[:len(tmp)-2])
			if err != nil {
				t.Fatalf("Bad indice computation: %v on line %d", err, lineNo)
			}
			message = make([][]byte, indice)
			genX = make([][]byte, indice)
			genY = make([][]byte, indice)
			continue
		}

		line = line[:len(line)-1]

		switch {
		case strings.HasPrefix(line, "Message = "):
			message[indiceMess], err = hex.DecodeString(line[10:])
			indiceMess++
		case strings.HasPrefix(line, "X = "):
			lineParsed = parseLine(line[4:])
			keyX, err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "Y = "):
			lineParsed = parseLine(line[4:])
			keyY, err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "GenX = "):
			lineParsed = parseLine(line[7:])
			genX[indiceGen], err = hex.DecodeString(lineParsed)
		case strings.HasPrefix(line, "GenY = "):
			lineParsed = parseLine(line[7:])
			genY[indiceGen], err = hex.DecodeString(lineParsed)
			indiceGen++
		case strings.HasPrefix(line, "Random = "):
			random, err = hex.DecodeString(line[9:])
		case strings.HasPrefix(line, "Commitment = "):
			indiceMess = 0
			indiceGen = 0
			commitment, err = hex.DecodeString(line[13:])

			var key ecdsa.PublicKey
			gen := make([]ecdsa.PublicKey, indice)
			key.X = new(big.Int).SetBytes(keyX)
			if err != nil {
				t.Fatalf("Error during commitment parsing at line: %d: %v", lineNo, err)
			}
			key.Y = new(big.Int).SetBytes(keyY)
			if err != nil {
				t.Fatalf("Error during commitment parsing at line: %d: %v", lineNo, err)
			}
			for i := 0; i < indice; i++ {
				gen[i].X = new(big.Int).SetBytes(genX[i])
				gen[i].Y = new(big.Int).SetBytes(genY[i])
				gen[i].Curve = crypto.S256()
			}
			key.Curve = crypto.S256()

			computedCommitment, err := Commit(message, &key, gen, random)
			if err != nil {
				t.Fatalf("Error during commitment computation at line: %d: %v", lineNo, err)
			}

			if bytes.Compare(commitment, computedCommitment) != 0 {
				t.Fatalf("Bad commitment at line: %d\n Got: %x\n Want:%x", lineNo, computedCommitment, commitment)
			}
		default:
		}
	}
}
