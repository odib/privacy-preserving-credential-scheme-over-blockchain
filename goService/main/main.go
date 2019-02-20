package main

import (
	"log"
	"net/http"

	"apipoc"

	"github.com/gorilla/mux"
)

// main function to boot up everything
func main() {
	apipoc.Init()
	router := mux.NewRouter()
	router.HandleFunc("/people", apipoc.GetPeople).Methods("GET")
	router.HandleFunc("/people/{id}", apipoc.GetPerson).Methods("GET")
	router.HandleFunc("/people/{id}", apipoc.CreatePerson).Methods("POST")
	router.HandleFunc("/people/{id}", apipoc.DeletePerson).Methods("DELETE")

	//return {"pub":"string", "priv":"string"}
	router.HandleFunc("/user/generateKey", apipoc.GenerateKey).Methods("GET")

	//input {"pub":"string", "value":int}
	router.HandleFunc("/user/commitment", apipoc.Commitment).Methods("POST")

	//input {"commitment":"string", "priv":"string"} priv is the private key of the IV
	router.HandleFunc("/iv/signCommitment", apipoc.SignCommitment).Methods("POST")

	//input {"s": string, "r": string, "pub": string, "commitment":string} pub is the public key of the IV
	//return {"verify":"true"} or {"verify":"false"}
	router.HandleFunc("/iv/verifySignature", apipoc.VerifySignature).Methods("POST")

	//input {"secret":"string", "pub":"string"}
	//return {"A":"string", "t": "string", "pubSecret":"string"}
	router.HandleFunc("/user/generateZKP/random", apipoc.GenerateZKPRandom).Methods("POST")

	//input {"secret":"string"}
	//return {"A":"string", "t": "string", "pubSecret":"string"}
	router.HandleFunc("/user/generateZKP/age", apipoc.GenerateZKPAge).Methods("POST")

	//input {"A":"string", "t":"string", "pub":"string", "pubSecret":"string"}
	//return {"verify":"true"} or {"verify":"false"}
	router.HandleFunc("/CP/verifyProof/random", apipoc.VerifyProofRandom).Methods("POST")

	//input {"A":"string", "t":"string", "pubSecret":"string"}
	//return {"verify":"true"} or {"verify":"false"}
	router.HandleFunc("/CP/verifyProof/age", apipoc.VerifyProofAge).Methods("POST")

	//input {"commitment":"string", "pubSecretAge":"string", "pubSecretRandom":"string"}
	//return {"verify":"true"} or {"verify":"false"}

	//return {"priv":"string", "g1Pub":"string" "g2Pub":"string"}
	router.HandleFunc("/user/generateKeyPairing", apipoc.GeneratePairingKey).Methods("GET")

	//input {"commitment":"string", "privCP":"string", "pubG2User":"string"}
	//return {"certificate":"string"} (if verification of some parameter fail, certificate is set to "false")
	router.HandleFunc("/CP/generateCertificate", apipoc.GenerateCertificate).Methods("POST")

	//input {"commitment":"string", "certificate":"string", "pubG1CP":"string", "pubG2User":"string"}
	//return {"verify":"true"} or {"verify":"false"}
	router.HandleFunc("/user/verifyCertificate", apipoc.VerifyCertificate).Methods("POST")

	//input {"commitment", "certificate", "pubG1CP", "pubG2User", "privUser"}
	//return {"blindCommitment", "blindCertificate", "blindPubG1CP", "blindPubG2User", "blindPrivUser", "blindGenerator", "blindFactor"}
	router.HandleFunc("/user/blindCertificate", apipoc.BlindCertificate).Methods("POST")

	//input {"blindCommitment", "blindPubG1CP", "blindPubG2User", "blindCertificate", "blindGenerator"}
	//return {"verify":"true"} or {"verify":"false"}
	router.HandleFunc("/SP/verifyBlindCertificate", apipoc.VerifyBlindedCertificate).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}
