package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
)

func f() {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}
	plainText := []byte("hello world")

	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, &privKey.PublicKey, plainText)
	if err != nil {
		log.Fatalf("could not encrypt data: %v", err)
	}
	log.Printf("%s\n", strconv.Quote(string(cipherText)))

	decryptedText, err := rsa.DecryptPKCS1v15(nil, privKey, cipherText)
	if err != nil {
		log.Fatalf("error decrypting cipher text: %v", err)
	}
	log.Printf("%s\n", decryptedText)

	hash := sha256.Sum256(plainText)
	fmt.Printf("The hash of my message is: %#x\n", hash)
	// The hash of my message is: 0xe6a8502561b8e2328b856b4dbe6a9448d2bf76f02b7820e5d5d4907ed2e6db80

	//用私钥在 hash 结果上生成签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	if err != nil {
		log.Fatalf("error creating signature: %v", err)
	}

	verify := func(pub *rsa.PublicKey, msg, signature []byte) error {
		hash := sha256.Sum256(msg)
		return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], signature)
	}
	fmt.Println(verify(&privKey.PublicKey, plainText, []byte("a bad signature")))
	// crypto/rsa: verification error
	fmt.Println(verify(&privKey.PublicKey, []byte("a different plain text"), signature))
	// crypto/rsa: verification error
	fmt.Println(verify(&privKey.PublicKey, plainText, signature))
	// <nil>

}
