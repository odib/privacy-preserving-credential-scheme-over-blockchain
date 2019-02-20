# Go Service in charge of Crypto material

## Requirements

- Golang
- ```go get github.com/ethereum/go-ethereum/crypto```
- ```go get github.com/gorilla/mux```
- ```go get golang.org/x/crypto/bn256```
- Add the src folder in your GOPATH environment variable

## Build

```go build``` inside the main folder

## Documentation

You can generate the api documentation using the command ```make gen-doc-docker```. It will create a ```doc``` folder and generate the documentation inside it.
