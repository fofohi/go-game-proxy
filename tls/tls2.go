package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

func main() {
	//生成一对新的公私钥
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}
	rootCertTmpl, err := CertTemplate()
	if err != nil {
		log.Fatalf("creating cert template: %v", err)
	}
	//在模板的基础上增加一些新的证书信息
	rootCertTmpl.IsCA = true //是否是CA
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	rootCert, rootCertPEM, err := CreateCert(rootCertTmpl, rootCertTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		log.Fatalf("error creating cert: %v", err)
	}
	fmt.Printf("%s\n", rootCertPEM)
	fmt.Printf("%#x\n", rootCert.Signature)
	// 将私钥用 pem 编码
	rootKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey),
	})
	fmt.Printf("rootKeyPEM :\n%s\n", rootKeyPEM)
	// 将证书和私钥都 pem 编码之后，结合起来生成最终的 TLS 证书
	rootTLSCert, err := tls.X509KeyPair(rootCertPEM, rootKeyPEM)
	if err != nil {
		log.Fatalf("invalid key pair: %v", err)
	}
	log.Println(rootTLSCert)
}

//证书模板，通过该模板默认设置一些证书需要的字段，比如序列号，组织信息，有效期等等
func CertTemplate() (*x509.Certificate, error) {
	//生成随机的序列号 (不同组织可以有不同的序列号生成方式)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())
	}
	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"szh, Inc."}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 36000), //1小时的有效期
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (cert *x509.Certificate, certPEM []byte, err error) {

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return
	}
	// parse the resulting certificate so we can use it again
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return
	}
	//将 certDER 用 pem 编码，生成 certPEM 证书
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	return
}
