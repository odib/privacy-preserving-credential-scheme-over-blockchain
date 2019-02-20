package apipoc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"converterhex"
	"cryptolib"

	"github.com/gorilla/mux"
)

var curve = elliptic.P256()

//Person is a struct containing the information of an individu
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

//Address is a struct containing the address of a person
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

//var people []Person{Person{ID: "1", Firstname: "John", Lastname: "Doe", Address: &api.Address{City: "City X", State: "State X"}, }
var people []Person

//Init initializes two persons
func Init() {
	people = append(people, Person{ID: "1", Firstname: "John", Lastname: "Doe", Address: &Address{City: "City X", State: "State X"}})
	people = append(people, Person{ID: "2", Firstname: "Koko", Lastname: "Doe", Address: &Address{City: "City Z", State: "State Y"}})
}

//GetPeople displays all from the people var
func GetPeople(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(people)
}

//GetPerson displays a single person data
func GetPerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

//CreatePerson creates a new person
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var person Person
	_ = json.NewDecoder(r.Body).Decode(&person)
	person.ID = params["id"]
	people = append(people, person)
	json.NewEncoder(w).Encode(people)
}

//DeletePerson deletes a person
func DeletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
			break
		}
		json.NewEncoder(w).Encode(people)
	}
}

/**
 * @api {get} /user/generateKey Generate key
 *
 * @apiName GenerateKey
 * @apiGroup User
 *
 * @apiDescription Return ECDSA private and public key for an user
 *
 * @apiSuccess {String} pub Return the publi key of the user
 * @apiSuccess {String} priv Return the private key of the user
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"pub": "012345...DEF",
 *			"priv": "72616e646f6d"
 *		}
 *
 */
func GenerateKey(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Return struct {
		Pub  string `json:"pub"`
		Priv string `json:"priv"`
	}
	priv, _ := ecdsa.GenerateKey(curve, rand.Reader)
	pubByte := elliptic.Marshal(priv.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	ret := Return{Pub: hex.EncodeToString(pubByte), Priv: priv.D.Text(16)}
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)

	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("GenerateKey: ", elapsed)
	return
}

//Pubic generator used to generate the ZKP and the commitment.
//This value is public and known by all the generator and verifier of proofs.
const (
	gx = "55126828932999412942359357557609819057251331021666655435427634109588796591969"
	gy = "63544379041443090754421963879511225825336925190710611664440074306356880601047"
)

/**
 * @api {post} /user/commitment Commitment Computing
 *
 * @apiName Commitment
 * @apiGroup User
 *
 * @apiDescription Return the Pedersen commitment for a given message
 *
 * @apiParam {String} pub ECDSA public key associated to the user (hexa decimal string)
 * @apiParam {String} value Value to be committed
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *		"pub": "04123456ABDE...",
 *		"value": "21"
 *	 }
 *
 * @apiSuccess {String} commitment Return the hexadecimal string representing the commitment of a given price and value.
 * @apiSuccess {String} random Return the haxadecimal string representing the random value used to compute the commitment.
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"commitment": "012345...DEF",
 *			"random": "72616e646f6d"
 *		}
 *
 */
func Commitment(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Pub string `json:"pub"`
		Age string `json:"age"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	pubByte, _ := converterhex.HexToByte(in.Pub)

	X, Y := elliptic.Unmarshal(curve, pubByte)
	pub := ecdsa.PublicKey{Curve: curve, X: X, Y: Y}

	gxInt, _ := new(big.Int).SetString(gx, 10)
	gyInt, _ := new(big.Int).SetString(gy, 10)

	gen := ecdsa.PublicKey{Curve: curve, X: gxInt, Y: gyInt}
	random := make([]byte, 16)
	rand.Read(random)
	commit, err := cryptolib.Commit([][]byte{[]byte(in.Age)}, &pub, []ecdsa.PublicKey{gen}, random)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Ret struct {
		Commitment string `json:"commitment"`
		Random     string `json:"random"`
	}
	res := Ret{Commitment: hex.EncodeToString(commit), Random: hex.EncodeToString(random)}
	retByte, _ := json.Marshal(res)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("Commitment: ", elapsed)
	return
}

/**
 * @api {post} /iv/signCommitment Sign a commitment
 *
 * @apiName SignCommitment
 * @apiGroup IV
 *
 * @apiDescription Return the ECDSA signature of the commitment
 *
 * @apiParam {String} commitment Commitment to be signed
 * @apiParam {String} priv Private key of the signer
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *		"commitment": "04123456ABDE...",
 *		"priv": "21ADC22..."
 *	 }
 *
 * @apiSuccess {String} r First member of the signature
 * @apiSuccess {String} s Second member of the signature (standard notation for signature see https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm)
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"r": "012345...DEF",
 *			"s": "7302616DEe6AA46f6d..."
 *		}
 *
 */
func SignCommitment(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Commitment string `json:"commitment"`
		Priv       string `json:"priv"`
		Pub        string `json:"pub"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	commit, _ := converterhex.HexToByte(in.Commitment)
	pubByte, _ := converterhex.HexToByte(in.Pub)
	x, y := elliptic.Unmarshal(curve, pubByte)
	pubKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}
	privD, _ := new(big.Int).SetString(in.Priv, 16)
	privKey := ecdsa.PrivateKey{PublicKey: pubKey, D: privD}

	h := sha256.New()
	h.Write(commit)
	hash := h.Sum(nil)

	rSign, sSign, _ := ecdsa.Sign(rand.Reader, &privKey, hash)

	type Ret struct {
		R string `json:"r"`
		S string `json:"s"`
	}
	ret := Ret{R: rSign.Text(16), S: sSign.Text(16)}
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("SignCommitment: ", elapsed)
	return

}

/**
 * @api {post} /iv/verifySignature Verify an ECDSA signature
 *
 * @apiName VerifySignature
 * @apiGroup IV
 *
 * @apiDescription Return true if the signature is valid, false else
 *
 * @apiParam {String} commitment Commitment which is signed
 * @apiParam {String} r First member of the signature
 * @apiParam {String} s Second member of the signature
 * @apiParam {String} pub Public key of the signer
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *		"commitment": "04123456ABDE...",
 *		"r": "21ADADAADC242...",
 *		"s": "2CCC1ADC22...",
 *		"pub": "04000242400FF21ADC22..."
 *	 }
 *
 * @apiSuccess {String} verify Return "true" if signature OK, "false" else
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"verify": "true"
 *		}
 *
 */
func VerifySignature(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		R          string `json:"r"`
		S          string `json:"s"`
		Commitment string `json:"commitment"`
		Pub        string `json:"pub"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	commit, _ := converterhex.HexToByte(in.Commitment)
	pubByte, _ := converterhex.HexToByte(in.Pub)
	x, y := elliptic.Unmarshal(curve, pubByte)
	pubKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}
	rInt, _ := new(big.Int).SetString(in.R, 16)
	sInt, _ := new(big.Int).SetString(in.S, 16)

	h := sha256.New()
	h.Write(commit)
	hash := h.Sum(nil)

	b := ecdsa.Verify(&pubKey, hash, rInt, sInt)
	type Ret struct {
		Verify string `json:"verify"`
	}
	var ret Ret
	if b == true {
		ret.Verify = "true"
	} else {
		ret.Verify = "false"
	}
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("VerifySignature: ", elapsed)
	return
}

/**
 * @api {post} /user/generateZKP/random Generate ZKP for the random
 *
 * @apiName GenerateZKPRandom
 * @apiGroup User
 *
 * @apiDescription Generate a ZKP for the random member of the commitment
 *
 * @apiParam {String} secret Secret to hide: random value used in commitment
 * @apiParam {String} pub Public key of the user
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *		"secret": "04123456ABDE...",
 *		"pub": "04000242400FF21ADC22..."
 *	 }
 *
 * @apiSuccess {String} A Random value used to compute and verify ZKP
 * @apiSuccess {String} t Public value containing the secret used to verify the ZKP
 * @apiSuccess {String} pubSecret Public generator value multiply by a secret which is secret*Generator
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"A": "01234ABC..."
 *	 		"t": "01234ABC...",
 *	 		"pubSecret": "01234ABC...",
 *		}
 *
 */
func GenerateZKPRandom(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Secret string `json:"secret"`
		Pub    string `json:"pub"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	pubByte, _ := converterhex.HexToByte(in.Pub)
	x, y := elliptic.Unmarshal(curve, pubByte)
	secret, _ := converterhex.HexToByte(in.Secret)

	random := make([]byte, 16)
	rand.Read(random)

	a, t := cryptolib.GenerateProof(curve, random, x, y, secret)
	pubSecretX, pubSecretY := curve.ScalarMult(x, y, secret)

	type Ret struct {
		A         string `json:"A"`
		T         string `json:"t"`
		PubSecret string `json:"pubSecret"`
	}
	var ret Ret
	AByte := elliptic.Marshal(curve, a.X, a.Y)
	ret.A = hex.EncodeToString(AByte)

	ret.T = hex.EncodeToString(t)
	ret.PubSecret = hex.EncodeToString(elliptic.Marshal(curve, pubSecretX, pubSecretY))

	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("GenerateZKP: ", elapsed)
	return
}

/**
 * @api {post} /user/generateZKP/age Generate ZKP for the committed value
 *
 * @apiName GenerateZKPAge
 * @apiGroup User
 *
 * @apiDescription Generate a ZKP for the secret member of the commitment
 *
 * @apiParam {String} secret Secret to hide: committed value
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *		"secret": "04123456ABDE...",
 *	 }
 *
 * @apiSuccess {String} A Random value used to compute and verify ZKP
 * @apiSuccess {String} t Public value containing the secret used to verify the ZKP
 * @apiSuccess {String} pubSecret Public value which is secret*G (a generator hard coded in the code. Must be public and know by the participant)
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"A": "01234ABC..."
 *	 		"t": "01234ABC...",
 *	 		"pubSecret": "01234ABC...",
 *		}
 *
 */
func GenerateZKPAge(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Secret string `json:"secret"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	x, _ := new(big.Int).SetString(gx, 10)
	y, _ := new(big.Int).SetString(gy, 10)
	secret := []byte(in.Secret)
	random := make([]byte, 16)
	rand.Read(random)
	a, t := cryptolib.GenerateProof(curve, random, x, y, secret)
	pubSecretX, pubSecretY := curve.ScalarMult(x, y, secret)

	type Ret struct {
		A         string `json:"A"`
		T         string `json:"t"`
		PubSecret string `json:"pubSecret"`
	}
	var ret Ret
	AByte := elliptic.Marshal(curve, a.X, a.Y)
	ret.A = hex.EncodeToString(AByte)

	ret.T = hex.EncodeToString(t)
	ret.PubSecret = hex.EncodeToString(elliptic.Marshal(curve, pubSecretX, pubSecretY))

	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("Generate ZKP Age: ", elapsed)
	return
}

/**
 * @api {post} /CP/verifyProof/random Verify the ZKP for the random value
 *
 * @apiName VerifyProofRandom
 * @apiGroup CP
 *
 * @apiDescription Verify the ZKP for the random member o the commitment
 *
 * @apiParam {String} A Random generated during the proof generation
 * @apiParam {String} t Public containing hidden generated during proof generation
 * @apiParam {String} pub Public key of the owner of the proof
 * @apiParam {String} pubSecret Public generator multiply by the secret value during the generation of the proof
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *	 		"A": "01234ABC...",
 *	 		"t": "01234ABC...",
 *			"pub": "0123EEADCD...",
 *	 		"pubSecret": "01234ABC...",
 *	 }
 *
 * @apiSuccess {String} verify Return "true" if proof is OK, "false" else
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"verify": "true"
 *		}
 *
 */
func VerifyProofRandom(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		A         string `json:"A"`
		T         string `json:"t"`
		Pub       string `json:"pub"`
		PubSecret string `json:"pubSecret"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	t, _ := converterhex.HexToByte(in.T)
	pubByte, _ := converterhex.HexToByte(in.Pub)
	x, y := elliptic.Unmarshal(curve, pubByte)
	pubKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	AByte, _ := converterhex.HexToByte(in.A)
	Ax, Ay := elliptic.Unmarshal(curve, AByte)
	AKey := ecdsa.PublicKey{Curve: curve, X: Ax, Y: Ay}

	pubSecretByte, _ := converterhex.HexToByte(in.PubSecret)
	xSecret, ySecret := elliptic.Unmarshal(curve, pubSecretByte)
	pubSecretKey := ecdsa.PublicKey{Curve: curve, X: xSecret, Y: ySecret}

	b := cryptolib.VerifyProof(curve, t, &AKey, &pubKey, &pubSecretKey)

	type Ret struct {
		Verify string `json:"verify"`
	}
	var ret Ret
	if b == true {
		ret.Verify = "true"
	} else {
		ret.Verify = "false"
	}
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("VerifyProofRandom: ", elapsed)
	return
}

/**
 * @api {post} /CP/verifyProof/age Verify the ZKP for the committed
 *
 * @apiName VerifyProofAge
 * @apiGroup CP
 *
 * @apiDescription Verify the ZKP for the committed value
 *
 * @apiParam {String} A Random generated during the proof generation
 * @apiParam {String} t Public containing hidden generated during proof generation
 * @apiParam {String} pubSecret Public generator multiply by the secret value during the generation of the proof
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *	 		"A": "01234ABC...",
 *	 		"t": "01234ABC...",
 *	 		"pubSecret": "01234ABC...",
 *	 }
 *
 * @apiSuccess {String} verify Return "true" if proof is OK, "false" else
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"verify": "true"
 *		}
 *
 */
func VerifyProofAge(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		A         string `json:"A"`
		T         string `json:"t"`
		PubSecret string `json:"pubSecret"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	t, _ := converterhex.HexToByte(in.T)
	x, _ := new(big.Int).SetString(gx, 10)
	y, _ := new(big.Int).SetString(gy, 10)
	pubKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	AByte, _ := converterhex.HexToByte(in.A)
	Ax, Ay := elliptic.Unmarshal(curve, AByte)
	AKey := ecdsa.PublicKey{Curve: curve, X: Ax, Y: Ay}

	pubSecretByte, _ := converterhex.HexToByte(in.PubSecret)
	xSecret, ySecret := elliptic.Unmarshal(curve, pubSecretByte)
	pubSecretKey := ecdsa.PublicKey{Curve: curve, X: xSecret, Y: ySecret}

	b := cryptolib.VerifyProof(curve, t, &AKey, &pubKey, &pubSecretKey)

	type Ret struct {
		Verify string `json:"verify"`
	}
	var ret Ret
	if b == true {
		ret.Verify = "true"
	} else {
		ret.Verify = "false"
	}
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("VerifyProofAge: ", elapsed)
	return
}

/**
 * @api {get} /User/generateKeyPairing Generate priv and pub pairing keys
 *
 * @apiName GeneratePairingKey
 * @apiGroup User
 *
 * @apiDescription Generate a private key and two public keys. Pairing goes from G1 x G2 -> GT. We need a public key for G1 and another for G2
 *
 * @apiSuccess {String} priv Private pairing key
 * @apiSuccess {String} g1Pub Public key for the first member
 * @apiSuccess {String} g2Pub Public key for the seconde member
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"priv": "01234ABC...",
 *	 		"g1Pub": "01234ABC...",
 *	 		"g2Pub": "01234ABC...",
 *		}
 *
 */
func GeneratePairingKey(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Ret struct {
		Priv  string `json:"priv"`
		G1Pub string `json:"g1Pub"`
		G2Pub string `json:"g2Pub"`
	}

	privByte, g1PubByte, g2PubByte, _ := cryptolib.GeneratePairingKey()

	ret := Ret{Priv: hex.EncodeToString(privByte), G1Pub: hex.EncodeToString(g1PubByte), G2Pub: hex.EncodeToString(g2PubByte)}
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("GeneratePairingKey: ", elapsed)
	return
}

/**
 * @api {post} /user/commitment Commitment Verification
 *
 * @apiName Commitment
 * @apiGroup User
 *
 * @apiDescription Verify that the commited values has been used in the commitment construction.
 *
 * @apiParam {String} pubSecretAge commited age value
 * @apiParam {String} pubSecretRandom commited random value
 * @apiParam {String} commitment commitment value
 *
 * @apiParamExample {json} Request-Example:
 *	{
 *
 *		"pubSecretAge": "27",
 *		"pubSecretRandom": "2390913AD...",
 *		"commitment": "0123456789ABC...",
 *	}
 *
 * @apiSuccess {String} verify true if the commitment is well construct, false else
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *			"verify": "true"
 *		}
 *
 */
func VerifyCommitment(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		PubSecretAge    string `json:"pubSecretAge"`
		PubSecretRandom string `json:"pubSecretRandom"`
		Commitment      string `json:"commitment"`
	}

	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	commit, _ := converterhex.HexToByte(in.Commitment)

	pubSecretAgeByte, _ := converterhex.HexToByte(in.PubSecretAge)
	xSecretAge, ySecretAge := elliptic.Unmarshal(curve, pubSecretAgeByte)
	pubSecretAgeKey := ecdsa.PublicKey{Curve: curve, X: xSecretAge, Y: ySecretAge}

	pubSecretRandomByte, _ := converterhex.HexToByte(in.PubSecretRandom)
	xSecretRandom, ySecretRandom := elliptic.Unmarshal(curve, pubSecretRandomByte)
	pubSecretRandomKey := ecdsa.PublicKey{Curve: curve, X: xSecretRandom, Y: ySecretRandom}

	commitX, commitY := elliptic.Unmarshal(curve, commit)

	tmpX, tmpY := curve.Add(pubSecretAgeKey.X, pubSecretRandomKey.X, pubSecretAgeKey.Y, pubSecretRandomKey.Y)

	type Ret struct {
		Verified string `json:"verify"`
	}
	var ret Ret

	if (tmpX.Cmp(commitX) != 0) || (tmpY.Cmp(commitY) != 0) {
		ret.Verified = "false"
		retByte, _ := json.Marshal(ret)
		w.Header().Set("content-type", "application/json")
		w.Write(retByte)

		return
	}
	ret.Verified = "true"
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)

	return
}

/**
 * @api {post} /CP/generateCertificate Generate a certificate
 *
 * @apiName GenerateCertificate
 * @apiGroup CP
 *
 * @apiDescription Generate a certificate used in the protocol
 *
 * @apiParam {String} commitment The committed attribute of the user
 * @apiParam {String} privCP The private pairing key of the certificate provider, used to compute the certificate
 * @apiParam {String} pubG2User The public key of the user. Second member public key is used
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *	 		"commitment": "01234ABC...",
 *	 		"privCP": "01234ABC...",
 *	 		"pubG2User": "01234ABC...",
 *	 }
 *
 * @apiSuccess {String} certificate Return the certificate of the user for the committed attributes
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"certificate": "1929422ABE"
 *		}
 *
 */
func GenerateCertificate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Commitment string `json:"commitment"`
		PubG2      string `json:"pubG2User"`
		PrivCP     string `json:"privCP"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	commit, _ := converterhex.HexToByte(in.Commitment)
	pubByte, _ := converterhex.HexToByte(in.PubG2)
	priv, _ := converterhex.HexToByte(in.PrivCP)

	type Ret struct {
		Certificate string `json:"certificate"`
	}
	var ret Ret

	cert, err := cryptolib.GenerateCertificate(commit, priv, pubByte)
	if err != nil {
		ret.Certificate = "false"
		certByte, _ := json.Marshal(ret)
		w.Header().Set("content-type", "application/json")
		w.Write(certByte)
		return
	}

	ret.Certificate = hex.EncodeToString(cert)
	certByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(certByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("GenerateCertificate: ", elapsed)
	return
}

/**
 * @api {post} /user/verifyCertificate Verify a certificate a certificate
 *
 * @apiName VerifyCertificate
 * @apiGroup User
 *
 * @apiDescription Verify a classic certificate with the public parameter. Used by the user to check if the CP is honest
 *
 * @apiParam {String} commitment The committed attribute of the user
 * @apiParam {String} certificate The certificate
 * @apiParam {String} pubG1CP The public key of the cetificate provider. First member is used
 * @apiParam {String} pubG2User The public key of the user. Second member public key is used
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *	 		"commitment": "01234ABC...",
 *	 		"certificate": "01234ABC...",
 *	 		"pubG1CP": "01234ABC...",
 *	 		"pubG2User": "01234ABC...",
 *	 }
 *
 * @apiSuccess {String} verify Return "true" if OK, "false" else
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"verify": "true"
 *		}
 *
 */
func VerifyCertificate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Commitment  string `json:"commitment"`
		Certificate string `json:"certificate"`
		PubG1CP     string `json:"pubG1CP"`
		PubG2User   string `json:"pubG2User"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)
	commit, _ := converterhex.HexToByte(in.Commitment)
	certificate, _ := converterhex.HexToByte(in.Certificate)
	pubG1CP, _ := converterhex.HexToByte(in.PubG1CP)
	pubG2User, _ := converterhex.HexToByte(in.PubG2User)

	type Ret struct {
		Verify string `json:"verify"`
	}
	var ret Ret
	b, err := cryptolib.VerifyCertificate(commit, certificate, pubG1CP, pubG2User)
	if err != nil || b != true {
		fmt.Println(err)
		ret.Verify = "false"
		retByte, _ := json.Marshal(ret)
		w.Header().Set("content-type", "application/json")
		w.Write(retByte)
		return
	}
	ret.Verify = "true"
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("VerifyCertificate: ", elapsed)
	return
}

/**
 * @api {post} /user/blindCertificate Blind certificate
 *
 * @apiName BlindCertificate
 * @apiGroup User
 *
 * @apiDescription Blind the certificate, pubG2User, the hashed commitment, pubG1CP, G1 (the generator of the curve)
 *
 * @apiParam {String} commitment The committed attribute of the user. The commitment is hashed inside the function
 * @apiParam {String} certificate The certificate
 * @apiParam {String} pubG1CP The public key of the cetificate provider. First member is used
 * @apiParam {String} pubG2User The public key of the user. Second member public key is used
 * @apiParam {String} privUser The pairing private key of the user
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *	 		"commitment": "01234ABC...",
 *	 		"certificate": "01234ABC...",
 *	 		"pubG1CP": "01234ABC...",
 *	 		"pubG2User": "01234ABC...",
 *	 		"privUser": "01234ABC...",
 *	 }
 *
 * @apiSuccess {String} blindCommitment The blinded commitment
 * @apiSuccess {String} blindCertificate The blinded certificate
 * @apiSuccess {String} blindPubG1CP The blinded public key G1 of the CP
 * @apiSuccess {String} blindPubG2User The blinded public key G2 o the user
 * @apiSuccess {String} blindPrivUser The blinded private key of ther user
 * @apiSuccess {String} blindGenerator The blinded G1 generator
 * @apiSuccess {String} blindFactor The random b which blind all the other values
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *	 		"blindCommitment": "01234ABC...",
 *	 		"blindCertificate": "01234ABC...",
 *	 		"blindPubG1CP": "01234ABC...",
 *	 		"blindPubG2User": "01234ABC...",
 *	 		"blindPrivUser": "01234ABC...",
 *			"blindGenerator": "BBBAABA11...",
 *			"blindFactor": "ABABAB113EE...",
 *		}
 *
 */
func BlindCertificate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		Commitment  string `json:"commitment"`
		Certificate string `json:"certificate"`
		PubG1CP     string `json:"pubG1CP"`
		PubG2User   string `json:"pubG2User"`
		PrivUser    string `json:"privUser"`
	}

	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)

	commit, _ := converterhex.HexToByte(in.Commitment)
	certificate, _ := converterhex.HexToByte(in.Certificate)
	pubG1CP, _ := converterhex.HexToByte(in.PubG1CP)
	pubG2User, _ := converterhex.HexToByte(in.PubG2User)
	privUser, _ := converterhex.HexToByte(in.PrivUser)

	commitByte, certificateByte, pubG1CPByte, pubG2UserByte, privUserByte, generatorByte, randomByte :=
		cryptolib.BlindCertificate(commit, certificate, pubG1CP, pubG2User, privUser)

	type Ret struct {
		Commitment  string `json:"blindCommitment"`
		Certificate string `json:"blindCertificate"`
		PubG1CP     string `json:"blindPubG1CP"`
		PubG2User   string `json:"blindPubG2User"`
		PrivUser    string `json:"blindPrivUser"`
		Generator   string `json:"blindGenerator"`
		Random      string `json:"blindFactor"`
	}

	ret := Ret{Commitment: hex.EncodeToString(commitByte), Certificate: hex.EncodeToString(certificateByte), PubG1CP: hex.EncodeToString(pubG1CPByte),
		PubG2User: hex.EncodeToString(pubG2UserByte), PrivUser: hex.EncodeToString(privUserByte),
		Generator: hex.EncodeToString(generatorByte), Random: hex.EncodeToString(randomByte)}

	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("BlindCertificate: ", elapsed)
	return
}

/**
 * @api {post} /SP/verifyBlindCertificate Blind certificate
 *
 * @apiName VerifyBlindedCertificate
 * @apiGroup SP
 *
 * @apiDescription Verify the blinded certificate with the  public parameter
 *
 * @apiParam {String} blindCommitment The blinded commitment
 * @apiParam {String} blindCertificate The blinded certificate
 * @apiParam {String} blindPubG1CP The blinded public key G1 of the CP
 * @apiParam {String} blindPubG2User The blinded public key G2 o the user
 * @apiParam {String} blindGenerator The blinded G1 generator
 *
 * @apiParamExample {json} Request-Example:
 *   {
 *	 		"blindCommitment": "01234ABC...",
 *	 		"blindCertificate": "01234ABC...",
 *	 		"blindPubG1CP": "01234ABC...",
 *	 		"blindPubG2User": "01234ABC...",
 *			"blindGenerator": "BBBAABA11...",
 *		}
 *	 }
 *
 * @apiSuccess {String} verify Return "true" if OK, "false" else
 *
 * @apiSuccessExample Success-Response:
 *	HTTP/1.1 200 OK
 *		{
 *			"verify": "true"
 *		}
 *
 */
func VerifyBlindedCertificate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	type Input struct {
		BlindCommitment  string `json:"blindCommitment"`
		BlindPubG1CP     string `json:"blindPubG1CP"`
		BlindPubG2User   string `json:"blindPubG2User"`
		BlindCertificate string `json:"blindCertificate"`
		BlindGenerator   string `json:"blindGenerator"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	var in Input
	json.Unmarshal(body, &in)

	blindCommit, _ := converterhex.HexToByte(in.BlindCommitment)
	blindCertificate, _ := converterhex.HexToByte(in.BlindCertificate)
	blindPubG1CP, _ := converterhex.HexToByte(in.BlindPubG1CP)
	blindPubG2User, _ := converterhex.HexToByte(in.BlindPubG2User)
	blindGenerator, _ := converterhex.HexToByte(in.BlindGenerator)

	b, err := cryptolib.VerifyBlindCertificate(blindCommit, blindCertificate, blindPubG1CP, blindPubG2User, blindGenerator)
	type Ret struct {
		Verify string `json:"verify"`
	}
	var ret Ret
	if err != nil || b != true {
		fmt.Println(err)
		ret.Verify = "false"
		retByte, _ := json.Marshal(ret)
		w.Header().Set("content-type", "application/json")
		w.Write(retByte)
		return
	}
	ret.Verify = "true"
	retByte, _ := json.Marshal(ret)
	w.Header().Set("content-type", "application/json")
	w.Write(retByte)
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println("VerifyBlindCertificate: ", elapsed)
	return
}
