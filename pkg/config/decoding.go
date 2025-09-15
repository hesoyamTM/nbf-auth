package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

//	func Decode(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
//		block, _ := pem.Decode([]byte(pemEncoded))
//		x509Encoded := block.Bytes
//		privateKey, _ := x509.ParseECPrivateKey(x509Encoded)
//
//		blockPub, _ := pem.Decode([]byte(pemEncodedPub))
//		x509EncodedPub := blockPub.Bytes
//		genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
//		publicKey := genericPublicKey.(*ecdsa.PublicKey)
//
//		return privateKey, publicKey
//	}

func DecodePrivateKey(pemEncoded string) (*ecdsa.PrivateKey, error) {
	const op = "config.DecodePrivateKey"

	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return privateKey, nil
}

func DecodePublicKey(pemEncodedPub string) (*ecdsa.PublicKey, error) {
	const op = "config.DecodePublicKey"

	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey, nil
}
