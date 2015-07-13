package commands

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "crypto/x509/pkix"
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
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// NewCertCmd is the subsilent root command.
var NewCertCmd = &cobra.Command{
    Use:   "new-cert",
    Short: "new-cert creates a new certificate",
    Long:  ``,
    Run: func(cmd *cobra.Command, args []string) {

        // Create logger
        writer := log.NewConcurrentWriter(os.Stdout)
        logger := log.NewLogger(writer, "new-cert")

        err := InitializeConfig(writer)
        if err != nil {
            return
        }

        // Setup directory structure
        if err := CreatePkiDirectories(logger, "."); err != nil {
            return
        }

        // generate private key
        privatekey, err := rsa.GenerateKey(rand.Reader, viper.GetInt("Bits"))
        if err != nil {
            logger.Warn("Error generating private key")
            return
        }

        // Create Certificate request
        csr, req, err := CreateCertificateRequest(logger, privatekey)
        if err != nil {
            logger.Warn("Error creating CA", "err", err)
            return
        }

        // Create Certificate
        crt, err := CreateCertificate(logger, csr, privatekey)
        if err != nil {
            logger.Warn("Error creating certificate", "err", err)
            return
        }

        // Save cert request
        pki := path.Join(".", "pki")
        SaveCertificateRequest(logger, req, path.Join(pki, "reqs", viper.GetString("Name")+".req"))

        // Save private key
        SavePrivateKey(logger, privatekey, path.Join(pki, "private", viper.GetString("Name")+".key"))

        // Save certificate
        SaveCertificate(logger, crt, path.Join(pki, "public", viper.GetString("Name")+".crt"))

    },
}

// Pointer to NewCertCmd used in initialization
var newCertCmd *cobra.Command

// Command line args
var (
    Name string
)

func init() {

    NewCertCmd.PersistentFlags().IntVarP(&KeyBits, "bits", "", 4096, "Number of bits in key")
    NewCertCmd.PersistentFlags().StringVarP(&Hosts, "hosts", "", "127.0.0.1", "IP of cert")
    NewCertCmd.PersistentFlags().IntVarP(&Years, "years", "", 10, "Number of years until the certificate expires")
    NewCertCmd.PersistentFlags().StringVarP(&Organization, "organization", "", "kappa-ca", "Organization for CA")
    NewCertCmd.PersistentFlags().StringVarP(&Country, "country", "", "USA", "Country of origin for CA")
    NewCertCmd.PersistentFlags().StringVarP(&Name, "name", "", "localhost", "Name of certificate")
    newCertCmd = NewCertCmd
}

// InitializeNewCertConfig sets up the command line options for creating a new certificate
func InitializeNewCertConfig(logger log.Logger) error {
    viper.SetDefault("Name", "localhost")

    if newCertCmd.PersistentFlags().Lookup("name").Changed {
        logger.Info("", "Name", Name)
        viper.Set("Name", Name)
    }

    return nil
}

// CreateCertificateRequest generates a new certificate request
func CreateCertificateRequest(logger log.Logger, key *rsa.PrivateKey) (*x509.CertificateRequest, []byte, error) {
    name, org, country := viper.GetString("Name"), viper.GetString("Organization"), viper.GetString("Country")

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
    hosts := strings.Split(viper.GetString("Hosts"), ",")
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

func ReadCertificate(filename string, certType string) (*pem.Block, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("could not read file")
    }

    pemBlock, _ := pem.Decode(data)
    if pemBlock == nil {
        return nil, fmt.Errorf("error decoding PEM format")
    } else if pemBlock.Type != certType {
        return nil, fmt.Errorf("error reading certificate: expected different header")
    }

    return pemBlock, nil
}

// CreateCertificate generates a new cert
func CreateCertificate(logger log.Logger, req *x509.CertificateRequest, key *rsa.PrivateKey) ([]byte, error) {
    years := viper.GetInt("Years")

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
        return nil, fmt.Errorf("failed to generate serial number: %s", err)
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
    hosts := strings.Split(viper.GetString("Hosts"), ",")
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
