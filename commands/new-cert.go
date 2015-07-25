package commands

import (
    "crypto/rand"
    "crypto/rsa"
    "os"
    "path"

    log "github.com/mgutz/logxi/v1"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "github.com/subsilent/kappa/auth"
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
        if err := auth.CreatePkiDirectories(logger, "."); err != nil {
            return
        }

        // generate private key
        privatekey, err := rsa.GenerateKey(rand.Reader, viper.GetInt("Bits"))
        if err != nil {
            logger.Warn("Error generating private key")
            return
        }

        // Create Certificate request
        csr, req, err := auth.CreateCertificateRequest(logger, privatekey,
            viper.GetString("Name"), viper.GetString("Organization"),
            viper.GetString("Country"), viper.GetString("Hosts"))
        if err != nil {
            logger.Warn("Error creating CA", "err", err)
            return
        }

        // Create Certificate
        crt, err := auth.CreateCertificate(logger, csr, privatekey,
            viper.GetInt("Years"), viper.GetString("Hosts"))
        if err != nil {
            logger.Warn("Error creating certificate", "err", err)
            return
        }

        // Save cert request
        pki := path.Join(".", "pki")
        auth.SaveCertificateRequest(logger, req, path.Join(pki, "reqs", viper.GetString("Name")+".req"))

        // Save private key
        auth.SavePrivateKey(logger, privatekey, path.Join(pki, "private", viper.GetString("Name")+".key"))

        // Save certificate
        auth.SaveCertificate(logger, crt, path.Join(pki, "public", viper.GetString("Name")+".crt"))

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
