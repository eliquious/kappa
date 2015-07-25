package commands

import (
    "crypto/rand"
    "crypto/rsa"
    "os"
    "path"

    "github.com/go-errors/errors"

    log "github.com/mgutz/logxi/v1"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// InitCACmd is the subsilent root command.
var InitCACmd = &cobra.Command{
    Use:   "init-ca",
    Short: "init-ca creates a new certificate authority",
    Long:  ``,
    Run: func(cmd *cobra.Command, args []string) {

        // Create logger
        writer := log.NewConcurrentWriter(os.Stdout)
        logger := log.NewLogger(writer, "init-ca")

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

        // Create CA
        cert, err := CreateCertificateAuthority(logger, privatekey)
        if err != nil {
            logger.Warn("Error creating CA", "err", errors.Wrap(err, 5))
            return
        }

        // Save cert
        pki := path.Join(".", "pki")
        SaveCertificate(logger, cert, path.Join(pki, "ca.crt"))

        // Save private key
        SavePrivateKey(logger, privatekey, path.Join(pki, "private", "ca.key"))

    },
}

// Pointer to InitCACmd used in initialization
var initCmd *cobra.Command

// Command line args
var (
    KeyBits      int
    Years        int
    Organization string
    Country      string
    Hosts        string
)

func init() {

    InitCACmd.PersistentFlags().IntVarP(&KeyBits, "bits", "", 4096, "Number of bits in key")
    InitCACmd.PersistentFlags().IntVarP(&Years, "years", "", 10, "Number of years until the CA certificate expires")
    InitCACmd.PersistentFlags().StringVarP(&Organization, "organization", "", "kappa-ca", "Organization for CA")
    InitCACmd.PersistentFlags().StringVarP(&Country, "country", "", "USA", "Country of origin for CA")
    InitCACmd.PersistentFlags().StringVarP(&Hosts, "hosts", "", "127.0.0.1", "Comma delimited list of IPs or domains")
    initCmd = InitCACmd
}

// InitializeCertAuthConfig sets up the command line options for creating a CA
func InitializeCertAuthConfig(logger log.Logger) error {

    viper.SetDefault("Bits", "4096")
    viper.SetDefault("Years", "10")
    viper.SetDefault("Organization", "kappa-ca")
    viper.SetDefault("Country", "USA")

    if initCmd.PersistentFlags().Lookup("bits").Changed {
        logger.Info("", "Bits", KeyBits)
        viper.Set("Bits", KeyBits)
    }
    if initCmd.PersistentFlags().Lookup("years").Changed {
        logger.Info("", "Years", Years)
        viper.Set("Years", Years)
    }
    if initCmd.PersistentFlags().Lookup("organization").Changed {
        logger.Info("", "Organization", Organization)
        viper.Set("Organization", Organization)
    }
    if initCmd.PersistentFlags().Lookup("country").Changed {
        logger.Info("", "Country", Country)
        viper.Set("Country", Country)
    }
    if initCmd.PersistentFlags().Lookup("hosts").Changed {
        logger.Info("", "Hosts", Hosts)
        viper.Set("Hosts", Hosts)
    }

    return nil
}

// // CreatePkiDirectories creates the directory structures for storing the public and private keys.
// func CreatePkiDirectories(logger log.Logger, root string) error {
//     pki := path.Join(root, "pki")

//     // Create pki directory
//     if err := os.MkdirAll(pki, os.ModeDir|0755); err != nil {
//         logger.Warn("Could not create pki/ directory", "err", err)
//         return err
//     }

//     // Create public directory
//     if err := os.MkdirAll(path.Join(pki, "public"), os.ModeDir|0755); err != nil {
//         logger.Warn("Could not create pki/public/ directory", "err", err)
//         return err
//     }

//     // Create private directory
//     if err := os.MkdirAll(path.Join(pki, "private"), os.ModeDir|0755); err != nil {
//         logger.Warn("Could not create pki/private/ directory", "err", err)
//         return err
//     }

//     // Create reqs directory
//     if err := os.MkdirAll(path.Join(pki, "reqs"), os.ModeDir|0755); err != nil {
//         logger.Warn("Could not create pki/reqs/ directory", "err", err)
//         return err
//     }

//     return nil
// }

// // CreateCertificateAuthority generates a new CA
// func CreateCertificateAuthority(logger log.Logger, key *rsa.PrivateKey) ([]byte, error) {
//     years, org, country := viper.GetInt("Years"), viper.GetString("Organization"), viper.GetString("Country")

//     // Generate subject key id
//     logger.Info("Generating SubjectKeyID")
//     subjectKeyID, err := GenerateSubjectKeyID(key)
//     if err != nil {
//         return nil, err
//     }

//     // Create serial number
//     logger.Info("Generating Serial Number")
//     serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
//     serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
//     if err != nil {
//         return nil, fmt.Errorf("failed to generate serial number: %s", err)
//     }

//     // Create template
//     logger.Info("Creating Certificate template")
//     template := &x509.Certificate{
//         IsCA: true,
//         BasicConstraintsValid: true,
//         SubjectKeyId:          subjectKeyID,
//         SerialNumber:          serialNumber,
//         Subject: pkix.Name{
//             Country:      []string{country},
//             Organization: []string{org},
//         },
//         PublicKeyAlgorithm: x509.RSA,
//         SignatureAlgorithm: x509.SHA512WithRSA,
//         NotBefore:          time.Now().Add(-600).UTC(),
//         NotAfter:           time.Now().AddDate(years, 0, 0).UTC(),

//         // see http://golang.org/pkg/crypto/x509/#KeyUsage
//         ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
//         KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
//     }

//     // Associate hosts
//     logger.Info("Adding Hosts to Certificate")
//     hosts := strings.Split(viper.GetString("Hosts"), ",")
//     for _, h := range hosts {
//         if ip := net.ParseIP(h); ip != nil {
//             template.IPAddresses = append(template.IPAddresses, ip)
//         } else {
//             template.DNSNames = append(template.DNSNames, h)
//         }
//     }

//     // Create cert
//     logger.Info("Generating Certificate")
//     cert, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
//     if err != nil {
//         return nil, err
//     }

//     return cert, nil
// }

// // GenerateSubjectKeyID creates a subject id based on a private key.
// func GenerateSubjectKeyID(key *rsa.PrivateKey) (bytes []byte, err error) {
//     pub, err := asn1.Marshal(key.PublicKey)
//     if err != nil {
//         return
//     }
//     hash := sha1.Sum(pub)
//     bytes = hash[:]
//     return
// }

// // SavePrivateKey saves a PrivateKey in the PEM format.
// func SavePrivateKey(logger log.Logger, key *rsa.PrivateKey, filename string) {
//     logger.Info("Saving Private Key")
//     pemfile, _ := os.Create(filename)
//     pemkey := &pem.Block{
//         Type:  "RSA PRIVATE KEY",
//         Bytes: x509.MarshalPKCS1PrivateKey(key)}
//     pem.Encode(pemfile, pemkey)
//     pemfile.Close()
// }

// // SavePublicKey saves a public key in the PEM format.
// func SavePublicKey(logger log.Logger, key *rsa.PrivateKey, filename string) {
//     logger.Info("Saving Public Key")
//     pemfile, _ := os.Create(filename)
//     bytes, _ := x509.MarshalPKIXPublicKey(key.PublicKey)

//     pemkey := &pem.Block{
//         Type:  "RSA PUBLIC KEY",
//         Bytes: bytes}
//     pem.Encode(pemfile, pemkey)
//     pemfile.Close()
// }

// // SaveCertificate saves a certificate in the PEM format.
// func SaveCertificate(logger log.Logger, cert []byte, filename string) {
//     logger.Info("Saving Certificate")
//     pemfile, _ := os.Create(filename)
//     pemkey := &pem.Block{
//         Type:  "CERTIFICATE",
//         Bytes: cert}
//     pem.Encode(pemfile, pemkey)
//     pemfile.Close()
// }
