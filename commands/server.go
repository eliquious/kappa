package commands

import (
    "io/ioutil"
    "os"
    "os/signal"
    "path"

    "github.com/eliquious/core"
    log "github.com/mgutz/logxi/v1"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/subsilent/kappa/ssh"
    crypto "golang.org/x/crypto/ssh"
)

// ServerCmd is the subsilent root command.
var ServerCmd = &cobra.Command{
    Use:   "server",
    Short: "server starts the database server",
    Long:  ``,
    Run: func(cmd *cobra.Command, args []string) {

        // Create logger
        writer := log.NewConcurrentWriter(os.Stdout)
        logger := log.NewLogger(writer, "subsilent")

        err := InitializeConfig(writer)
        if err != nil {
            return
        }

        // Create data directory
        if err := os.MkdirAll(viper.GetString("DataPath"), os.ModeDir|0655); err != nil {
            logger.Warn("Could not create data directory", "err", err)
            return
        }

        // Connect to database
        cwd, err := os.Getwd()
        if err != nil {
            logger.Error("Could not get working directory", "error", err)
            return
        }

        file := path.Join(cwd, viper.GetString("DataPath"), "meta.db")
        logger.Info("Connecting to database", "file", file)
        factory := core.BoltDatabaseFactory{file}
        conn, err := factory.Connect()
        if err != nil {
            logger.Error("Could not connect to database", "error", err)
            return
        }

        // Get SSH Key file
        sshKeyFile := viper.GetString("SSHKey")
        logger.Info("Reading private key", "file", sshKeyFile)

        // Read SSH Key
        keyBytes, err := ioutil.ReadFile(sshKeyFile)
        if err != nil {
            logger.Error("Private key could not be read", "error", err)
            return
        }

        // Get private key
        privateKey, err := crypto.ParsePrivateKey(keyBytes)
        if err != nil {
            logger.Error("Private key could not be parsed", "error", err)
            return
        }

        // Setup SSH Server
        sshLogger := log.NewLogger(writer, "ssh")

        sshServer, err := ssh.NewSSHServer(sshLogger, conn, privateKey)
        if err != nil {
            logger.Error("SSH Server could not be configured", "error", err)
            return
        }

        // Setup signal structures
        closer := make(chan bool)
        sshServer.Run(logger, closer)

        // Handle signals
        sig := make(chan os.Signal, 1)
        signal.Notify(sig, os.Interrupt, os.Kill)

        // Wait for signal
        logger.Info("Ready to serve requests")
        <-sig

        // Shut down SSH server
        logger.Info("Shutting down servers.")
        sshServer.Wait()
        <-closer
    },
}

// Pointer to ServerCmd used in initialization
var serverCmd *cobra.Command

// Command line args
var (
    SSHKey     string
    TLSCert    string
    TLSKey     string
    DataPath   string
    SSHListen  string
    HTTPListen string
)

func init() {

    ServerCmd.PersistentFlags().StringVarP(&SSHKey, "ssh-key", "", "", "Private key to identify server with")
    ServerCmd.PersistentFlags().StringVarP(&TLSCert, "tls-cert", "", "", "TLS certificate file")
    ServerCmd.PersistentFlags().StringVarP(&TLSKey, "tls-key", "", "", "TLS private key file")
    ServerCmd.PersistentFlags().StringVarP(&DataPath, "data", "D", "", "Data directory")
    ServerCmd.PersistentFlags().StringVarP(&SSHListen, "ssh-listen", "", "", "Host and port for SSH server to listen on")
    ServerCmd.PersistentFlags().StringVarP(&HTTPListen, "http-listen", "", ":", "Host and port for HTTP server to listen on")
    serverCmd = ServerCmd
}

// InitializeServerConfig sets up the config options for the database servers.
func InitializeServerConfig(logger log.Logger) error {

    // Load default settings
    logger.Info("Loading default server settings")
    viper.SetDefault("SSHKey", "ssh-identity")
    viper.SetDefault("TLSCert", "tls-identity.crt")
    viper.SetDefault("TLSKey", "tls-identity.key")
    viper.SetDefault("DataPath", "./data")
    viper.SetDefault("SSHListen", ":9022")
    viper.SetDefault("HTTPListen", ":19022")

    if serverCmd.PersistentFlags().Lookup("ssh-key").Changed {
        logger.Info("", "SSHKey", SSHKey)
        viper.Set("SSHKey", SSHKey)
    }
    if serverCmd.PersistentFlags().Lookup("tls-cert").Changed {
        logger.Info("", "TLSCert", TLSCert)
        viper.Set("TLSCert", TLSCert)
    }
    if serverCmd.PersistentFlags().Lookup("tls-key").Changed {
        logger.Info("", "TLSKey", TLSKey)
        viper.Set("TLSKey", TLSKey)
    }
    if serverCmd.PersistentFlags().Lookup("ssh-listen").Changed {
        logger.Info("", "SSHListen", SSHListen)
        viper.Set("SSHListen", SSHListen)
    }
    if serverCmd.PersistentFlags().Lookup("http-listen").Changed {
        logger.Info("", "HTTPListen", HTTPListen)
        viper.Set("HTTPListen", HTTPListen)
    }
    if serverCmd.PersistentFlags().Lookup("data").Changed {
        logger.Info("", "DataPath", DataPath)
        viper.Set("DataPath", DataPath)
    }

    return nil
}
