package certificate

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	secretHelper "github.com/santi1s/yak/internal/cmd/secret"
	"github.com/santi1s/yak/internal/helper"
	log "github.com/sirupsen/logrus"
)

// Get the private key from the secret
func getValueFromSecret(config *SecretConfig, key string, dataType string, version int) (interface{}, error) {
	name := config.Path
	env := config.Env
	platform := config.Platform

	log.Println("Get value for key '" + key + "' in secret platform:" + platform + ",env:" + env + ",path:" + name)

	// sre are the only users of this task for the moment
	vaultConfig, err := helper.GetVaultConfig(platform, env)
	if err != nil {
		return "", err
	}
	clients, err := helper.VaultLoginWithAwsAndGetClients(vaultConfig)
	if err != nil {
		return "", err
	}

	if env == "" {
		env = "common"
	}
	secretPath := path.Join(env, name)
	if version == 0 {
		version, err = secretHelper.GetLatestVersion(clients, secretPath)
		if err != nil {
			return "", err
		}
	}

	secret, err := secretHelper.ReadSecretData(clients, secretPath, version)
	if err != nil {
		return "", err
	}

	value := secret.Data[dataType].(map[string]interface{})[key]
	return value, nil
}

// generate the CSR file from the private key stored in the secret and the conf file in /sslcerts
func generateCSR(privateKey string, configFilePath string) (string, error) {
	log.Println("Generating the CSR")

	privateKeyFile, err := os.CreateTemp("", "server.key")
	if err != nil {
		return "", err
	}

	_, err = privateKeyFile.WriteString(privateKey)
	if err != nil {
		return "", err
	}

	defer os.Remove(privateKeyFile.Name())
	defer privateKeyFile.Close()

	csrFile, err := os.CreateTemp("", "server.csr")
	if err != nil {
		return "", err
	}

	defer os.Remove(csrFile.Name())
	defer csrFile.Close()

	// Generate the CSR
	csr, err := exec.Command("openssl", "req", "-new", "-sha256", "-key", privateKeyFile.Name(), "-config", configFilePath).CombinedOutput() //#nosec
	if err != nil {
		log.Errorln(string(csr))
		return "", err
	}

	_, err = csrFile.Write(csr)
	if err != nil {
		return "", err
	}

	log.Println("Checking CSR integrity")
	// Check CSR integrity
	out, err := exec.Command("openssl", "req", "-text", "-noout", "-verify", "-in", csrFile.Name()).CombinedOutput() //#nosec
	if err != nil {
		log.Errorln(string(out))
		return "", err
	}

	return string(csr), nil
}

func checkExpiration(certificate *CertificateType) error {
	log.Println("Checking certificate expiration is in +30 days")
	if certificate.Dates == nil {
		return errors.New("certificate has no expiration date, maybe not yet validated by Gandi")
	}
	endDate := certificate.Dates.EndsAt
	diff := time.Until(endDate)
	if diff < 30*24 {
		return errors.New("expiration date of certificate is in less than 30 days")
	}
	return nil
}

func parseCertificate(cert string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}
	return parsedCert, nil
}

// ExtractPublicKeyFromCertificate extracts the public key from a PEM-encoded certificate.
func ExtractPublicKeyFromCertificate(certPEM string) (string, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return "", errors.New("failed to parse certificate PEM")
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Encode the public key to PEM format
	pubBytes, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(pubPEM), nil
}

// ExtractPublicKeyFromPrivateKey extracts the public key from a PEM-encoded private key.
func ExtractPublicKeyFromPrivateKey(privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", errors.New("failed to parse private key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			privateKey, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return "", fmt.Errorf("failed to parse private key: %v", err)
			}
		}
	}

	var publicKey crypto.PublicKey
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		publicKey = key.Public()
	case *ecdsa.PrivateKey:
		publicKey = key.Public()
	case ed25519.PrivateKey:
		publicKey = key.Public()
	default:
		return "", errors.New("unsupported private key type")
	}

	// Encode the public key to PEM format
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(pubPEM), nil
}

func comparePublicKeys(privateKey string, certificate string) error {
	log.Println("Comparing publickey with private key")
	pubkeyFromCert, err := ExtractPublicKeyFromCertificate(certificate)
	if err != nil {
		return err
	}

	pubkeyFromPrivate, err := ExtractPublicKeyFromPrivateKey(privateKey)
	if err != nil {
		return err
	}
	if pubkeyFromCert != pubkeyFromPrivate {
		return errors.New("the private key does not match the certificate")
	}
	return nil
}

func compareCertificates(oldCertificate string, newCertificate string) bool {
	log.Println("Check that the certificate from the API is different from the secret")

	// Remove spaces, newlines and tabs
	oldCertificateStr := strings.ReplaceAll(oldCertificate, " ", "")
	oldCertificateStr = strings.ReplaceAll(oldCertificateStr, "\t", "")
	oldCertificateStr = strings.ReplaceAll(oldCertificateStr, "\n", "")

	newCertificateStr := strings.ReplaceAll(newCertificate, " ", "")
	newCertificateStr = strings.ReplaceAll(newCertificateStr, "\t", "")
	newCertificateStr = strings.ReplaceAll(newCertificateStr, "\n", "")

	if oldCertificateStr == newCertificateStr {
		log.Println("The certificates from the API and stored in the secret are identical.")
		return false
	}

	return true
}
