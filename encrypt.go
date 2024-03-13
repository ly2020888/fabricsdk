package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"time"
)

// 使用ECDSA私钥对数据进行签名
func signData(privateKey *ecdsa.PrivateKey, data []byte) (string, error) {
	hashed := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed[:])
	if err != nil {
		return "", err
	}

	// 将签名的r和s组合成一个字节切片
	signature := append(r.Bytes(), s.Bytes()...)
	return base64.StdEncoding.EncodeToString(signature), nil
}

// 使用ECDSA公钥验证签名
func verifySignature(publicKey *ecdsa.PublicKey, data []byte, signature string) error {
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	// 将签名字节切片拆分为r和s
	rBytes := signatureBytes[:len(signatureBytes)/2]
	sBytes := signatureBytes[len(signatureBytes)/2:]
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	hashed := sha256.Sum256(data)
	if ecdsa.Verify(publicKey, hashed[:], r, s) {
		return nil
	} else {
		return fmt.Errorf("signature verification failed")
	}
}

func SignWithTime(timestamp time.Time) (time.Time, string) {
	// 读取ECDSA私钥文件
	keyFile := "keystore/private-key.pem"
	privateKeyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatal("Failed to read private key file: ", err)
	}

	// 解析ECDSA私钥
	privateKey, err := parseECDSAPrivateKey(privateKeyBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	// 将时间戳转换为字节数组
	timestampBytes := []byte(fmt.Sprintf("%v", timestamp))

	// 使用ECDSA私钥签名时间戳
	signature, err := signData(privateKey, timestampBytes)
	if err != nil {
		log.Fatal("Failed to sign data: ", err)
	}

	fmt.Println("Original Timestamp:", timestamp)
	fmt.Println("Signature:", signature)

	// 使用ECDSA公钥验证签名
	err = verifySignature(&privateKey.PublicKey, timestampBytes, signature)
	if err != nil {
		log.Fatal("Signature verification failed: ", err)
	} else {
		fmt.Println("Signature verification succeeded.")
	}
	return timestamp, signature
}

// 用于解析ECDSA私钥的辅助函数
func parseECDSAPrivateKey(keyBytes []byte) (*ecdsa.PrivateKey, error) {
	privateKey, err := x509.ParseECPrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
