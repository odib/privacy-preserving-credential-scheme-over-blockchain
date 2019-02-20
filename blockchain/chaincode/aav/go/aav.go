/*
 This code is written by Omar DIB (IRT SystemX), and Clément HUYART (ERCOM)
 Last Modifcations: December 12 2019
*/

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cryptoFunc"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type aav struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	AavKey     string `json:"aavKey"`    //the fieldtags are needed to keep case from bouncing around
	AavVal     string `json:"aavVal"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initAav" { //create a new aav
		return t.initAav(stub, args)
	} else if function == "verify" { //start anonymous on chain verification
		return t.verify(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initAav - create a new Aav, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initAav(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1   
	// "aavKey", "aavVal",
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init aav chaincode")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	aavKey := args[0]
	aavVal := args[1]

	// ==== Check if aav already exists ====
	aavAsBytes, err := stub.GetState(aavKey)
	if err != nil {
		return shim.Error("Failed to get aav: " + err.Error())
	} else if aavAsBytes != nil {
		fmt.Println("This aav already exists: " + aavKey)
		return shim.Error("This aav exists: " + aavKey)
	}

	// ==== Create aav object and marshal to JSON ====
	objectType := "aav"
	aav := &aav{objectType, aavKey, aavVal}
	aavJSONasBytes, err := json.Marshal(aav)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save aav to state ===
	err = stub.PutState(aavKey, aavJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init aav")
	return shim.Success(nil)
}


// ================================================================================================
// verify: on chain anonymous attribute verifications
// This function will be executed from the java orchestrator module with the right parameters
// ============================================================================================

func (t *SimpleChaincode) verify(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	blindCommit, _ := hexToByte(args[0])
	blindCertificate, _ := hexToByte(args[1])
	blindPubG1CP, _ := hexToByte(args[2])
	blindPubG2User, _ := hexToByte(args[3])
	blindGenerator, _ := hexToByte(args[4])

	start := time.Now()
	b, err := cryptoFunc.VerifyBlindCertificate(blindCommit, blindCertificate, blindPubG1CP, blindPubG2User, blindGenerator)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("VerifyBlindSignature time: ", elapsed)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(b)
	return shim.Success(nil)
}

//HexToByte converts the string containing an hexa decimal value into a byte representation
func hexToByte(s string) ([]byte, error) {
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
