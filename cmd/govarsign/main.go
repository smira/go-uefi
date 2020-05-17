package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/foxboron/goefi/efi"
	"github.com/foxboron/goefi/efi/attributes"
	"github.com/foxboron/goefi/efi/signature"
)

func ReadKey(path string) *rsa.PrivateKey {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode(b)
	if block == nil {
		panic("failed to parsePEM block containg the public key!")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded private key: " + err.Error())
	}
	switch priv := priv.(type) {
	case *rsa.PrivateKey:
		return priv
	default:
		panic("unknown type of public key")
	}
}

func ReadCert(path string) *x509.Certificate {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode(b)
	if block == nil {
		panic("failed to parsePEM block containg the public key!")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	return cert
}

func main() {
	key := flag.String("key", "", "Key")
	cert := flag.String("cert", "", "Certificate")
	variable := flag.String("var", "", "variable")
	flag.Parse()
	args := flag.Args()
	if len(os.Args) == 1 {
		fmt.Println("govarsign: -key <key> -cert <cert> -var <variable> [input] [output]")
	}
	if len(os.Args) == 2 {
		fmt.Println("Missing input and output file")
		os.Exit(1)
	}
	b, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	attrs := efi.ValidAttributes[*variable]
	attrs |= attributes.EFI_VARIABLE_APPEND_WRITE

	ctx := &signature.SigningContext{
		Cert:    ReadCert(*cert),
		Key:     ReadKey(*key),
		Varname: []byte(*variable),
		Guid:    attributes.EFI_GLOBAL_VARIABLE,
		Attr:    attrs,
		Data:    b,
	}
	signedVariable := signature.NewSignedEFIVariable(ctx)

	buf := new(bytes.Buffer)
	signature.WriteEFIVariableAuthencation2(buf, *signedVariable)
	buf.Write(b)
	err = ioutil.WriteFile(args[1], buf.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
