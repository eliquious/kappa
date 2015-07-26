package auth

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/mgutz/logxi/v1"
	"golang.org/x/crypto/ssh"
)

// CreateFingerprint generates an md5 fingerprint
func CreateFingerprint(key []byte) string {
	// Hash key
	h := md5.New()
	h.Write(key)

	// Create Fingerprint
	var fingerprint string
	hexidecimal := hex.EncodeToString(h.Sum(nil))
	for i := 0; i < len(hexidecimal); i += 2 {
		fingerprint += hexidecimal[i : i+2]
		if i+2 < len(hexidecimal) {
			fingerprint += ":"
		}
	}
	return fingerprint
}

// CreatePkiDirectories creates the directory structures for storing the public and private keys.
func CreatePkiDirectories(logger log.Logger, root string) error {
	pki := path.Join(root, "pki")

	// Create pki directory
	if err := os.MkdirAll(pki, os.ModeDir|0755); err != nil {
		logger.Warn("Could not create pki/ directory", "err", err.Error())
		return err
	}

	// Create public directory
	if err := os.MkdirAll(path.Join(pki, "public"), os.ModeDir|0755); err != nil {
		logger.Warn("Could not create pki/public/ directory", "err", err.Error())
		return err
	}

	// Create private directory
	if err := os.MkdirAll(path.Join(pki, "private"), os.ModeDir|0755); err != nil {
		logger.Warn("Could not create pki/private/ directory", "err", err.Error())
		return err
	}

	// Create reqs directory
	if err := os.MkdirAll(path.Join(pki, "reqs"), os.ModeDir|0755); err != nil {
		logger.Warn("Could not create pki/reqs/ directory", "err", err.Error())
		return err
	}

	return nil
}

// CreateCertificateAuthority generates a new CA
func CreateCertificateAuthority(logger log.Logger, key *rsa.PrivateKey, years int, org, country, hostList string) ([]byte, error) {

	// Generate subject key id
	logger.Info("Generating SubjectKeyID")
	subjectKeyID, err := GenerateSubjectKeyID(key)
	if err != nil {
		return nil, err
	}

	// Create serial number
	logger.Info("Generating Serial Number")
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err.Error())
	}

	// Create template
	logger.Info("Creating Certificate template")
	template := &x509.Certificate{
		IsCA: true,
		BasicConstraintsValid: true,
		SubjectKeyId:          subjectKeyID,
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			Country:      []string{country},
			Organization: []string{org},
		},
		PublicKeyAlgorithm: x509.RSA,
		SignatureAlgorithm: x509.SHA512WithRSA,
		NotBefore:          time.Now().Add(-600).UTC(),
		NotAfter:           time.Now().AddDate(years, 0, 0).UTC(),

		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// Associate hosts
	logger.Info("Adding Hosts to Certificate")
	hosts := strings.Split(hostList, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Create cert
	logger.Info("Generating Certificate")
	cert, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GenerateSubjectKeyID creates a subject id based on a private key.
func GenerateSubjectKeyID(key *rsa.PrivateKey) (bytes []byte, err error) {
	pub, err := asn1.Marshal(key.PublicKey)
	if err != nil {
		return
	}
	hash := sha1.Sum(pub)
	bytes = hash[:]
	return
}

// SavePrivateKey saves a PrivateKey in the PEM format.
func SavePrivateKey(logger log.Logger, key *rsa.PrivateKey, filename string) {
	logger.Info("Saving Private Key")
	pemfile, _ := os.Create(filename)
	pemkey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key)}
	pem.Encode(pemfile, pemkey)
	pemfile.Close()
}

// SavePublicKey saves a public key in the PEM format.
func SavePublicKey(logger log.Logger, key *rsa.PrivateKey, filename string) {
	logger.Info("Saving Public Key")
	pemfile, _ := os.Create(filename)
	bytes, _ := x509.MarshalPKIXPublicKey(key.PublicKey)

	pemkey := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes}
	pem.Encode(pemfile, pemkey)
	pemfile.Close()
}

// SaveCertificate saves a certificate in the PEM format.
func SaveCertificate(logger log.Logger, cert []byte, filename string) {
	logger.Info("Saving Certificate")
	pemfile, _ := os.Create(filename)
	pemkey := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert}
	pem.Encode(pemfile, pemkey)
	pemfile.Close()
}

// CreateCertificateRequest generates a new certificate request
func CreateCertificateRequest(logger log.Logger, key *rsa.PrivateKey, name, org, country, hostList string) (*x509.CertificateRequest, []byte, error) {

	// Create template
	logger.Info("Creating Certificate template")
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{country},
			Organization:       []string{org},
			OrganizationalUnit: []string{name},
		},
	}

	// Associate hosts
	logger.Info("Adding Hosts to Certificate")
	hosts := strings.Split(hostList, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Create cert
	logger.Info("Generating Certificate")
	cert, err := x509.CreateCertificateRequest(rand.Reader, template, key)
	if err != nil {
		return nil, nil, err
	}

	return template, cert, nil
}

// SaveCertificateRequest saves a certificate in the PEM format.
func SaveCertificateRequest(logger log.Logger, cert []byte, filename string) {
	logger.Info("Saving Certificate Request")
	pemfile, _ := os.Create(filename)
	pemkey := &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: cert}
	pem.Encode(pemfile, pemkey)
	pemfile.Close()
}

// ReadCertificate reads a cert file and validates the header
func ReadCertificate(filename string, certType string) (*pem.Block, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file")
	}

	// Decode PEM file
	pemBlock, _ := pem.Decode(data)
	if pemBlock == nil {
		return nil, fmt.Errorf("error decoding PEM format")
	} else if pemBlock.Type != certType {
		return nil, fmt.Errorf("error reading certificate: expected different header")
	}

	return pemBlock, nil
}

// ReadPrivateKey reads a private key file
func ReadPrivateKey(logger log.Logger, keyFile string) (privateKey ssh.Signer, err error) {

	// Read SSH Key
	keyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		logger.Warn("Private key could not be read", "error", string(err.Error()))
		return
	}

	// Get private key
	privateKey, err = ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		logger.Warn("Private key could not be parsed", "error", err.Error())
	}
	return
}

// CreateCertificate generates a new cert
func CreateCertificate(logger log.Logger, req *x509.CertificateRequest, key *rsa.PrivateKey, years int, hostList string) ([]byte, error) {

	// Read CA
	logger.Info("Reading Certificate Authority")
	pemBlock, err := ReadCertificate(path.Join(".", "pki", "ca.crt"), "CERTIFICATE")
	if err != nil {
		return nil, err
	}

	// Decrypt PEM
	logger.Info("Decoding Certificate Authority Public Key")
	authority, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	logger.Info("Reading Certificate Authority Private Key")
	pemBlock, err = ReadCertificate(path.Join(".", "pki", "private", "ca.key"), "RSA PRIVATE KEY")
	if err != nil {
		return nil, err
	}

	logger.Info("Parsing Certificate Authority Private Key")
	priv, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// Generate subject key id
	logger.Info("Generating SubjectKeyID")
	subjectKeyID, err := GenerateSubjectKeyID(key)
	if err != nil {
		return nil, err
	}

	// Create serial number
	logger.Info("Generating Serial Number")
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err.Error())
	}

	// Create template
	logger.Info("Creating Certificate template")
	template := &x509.Certificate{
		IsCA: false,
		BasicConstraintsValid: false,
		SubjectKeyId:          subjectKeyID,
		SerialNumber:          serialNumber,
		Subject:               req.Subject,
		PublicKeyAlgorithm:    x509.RSA,
		SignatureAlgorithm:    x509.SHA512WithRSA,
		NotBefore:             time.Now().Add(-600).UTC(),
		NotAfter:              time.Now().AddDate(years, 0, 0).UTC(),

		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,

		UnknownExtKeyUsage: nil,

		// Subject Alternative Name
		DNSNames: nil,

		PermittedDNSDomainsCritical: false,
		PermittedDNSDomains:         nil,
	}

	// Associate hosts
	logger.Info("Adding Hosts to Certificate")
	hosts := strings.Split(hostList, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Create cert
	logger.Info("Generating Certificate")
	cert, err := x509.CreateCertificate(rand.Reader, template, authority, &key.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
