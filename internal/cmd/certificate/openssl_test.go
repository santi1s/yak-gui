package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateTestCSR(t *testing.T, key *rsa.PrivateKey) []byte {
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "example.com",
			Organization: []string{"Example Organization"},
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		t.Fatalf("error while creating CSR %v", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	return csrPEM
}

func TestGenerateCSR(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating private key for testing : %v", err)
	}

	key2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating private key for testing : %v", err)
	}

	csrPEM1 := generateTestCSR(t, key)
	csrPEM2 := generateTestCSR(t, key2)

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	confFile, err := os.CreateTemp("", "csr.conf")
	if err != nil {
		t.Fatal("error creating conf file")
	}

	_, err = confFile.WriteString("[req]\ndistinguished_name = req_distinguished_name\nx509_extensions = v3_req\nprompt = no\n[req_distinguished_name]\nC = IT\nST = Milan\nL = Milan\nO = Doctolib SRL\nOU = IT\nCN = *.doctolib.it\n[v3_req]\nkeyUsage = keyEncipherment, dataEncipherment\nextendedKeyUsage = serverAuth")
	if err != nil {
		t.Fatal("error writing in conf file")
	}
	defer os.Remove(confFile.Name())
	defer confFile.Close()

	csrPEMFromMethod, err := generateCSR(string(pem.EncodeToMemory(privateKeyPEM)), confFile.Name())
	if err != nil {
		t.Fatal("error calling generateCSR")
	}

	// Décoder les CSR PEM
	block1, _ := pem.Decode(csrPEM1)
	block2, _ := pem.Decode(csrPEM2)
	blockFromMethod, _ := pem.Decode([]byte(csrPEMFromMethod))

	if block1 == nil || block2 == nil || blockFromMethod == nil {
		t.Fatal("csr pem could not be decoded")
		return // Ensures the static linter no further code is executed
	}

	csr1, err := x509.ParseCertificateRequest(block1.Bytes)
	if err != nil {
		t.Fatalf("error while analysing csr 1 : %v", err)
	}

	csr2, err := x509.ParseCertificateRequest(block2.Bytes)
	if err != nil {
		t.Fatalf("error while analysing csr 2: %v", err)
	}

	csrFromMethod, err := x509.ParseCertificateRequest(blockFromMethod.Bytes)
	if err != nil {
		t.Fatalf("error while analysing csr 4: %v", err)
	}

	// Extraire les clés publiques des CSR
	pubKey1 := csr1.PublicKey.(*rsa.PublicKey)
	pubKey2 := csr2.PublicKey.(*rsa.PublicKey)
	pubKeyFromMethod := csrFromMethod.PublicKey.(*rsa.PublicKey)

	resKO := pubKey1.Equal(pubKey2)
	resOK := pubKey1.Equal(pubKeyFromMethod)
	assert.Equal(t, resOK, true)
	assert.Equal(t, resKO, false)
}

func TestCheckExpiration(t *testing.T) {
	oldDate := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	today := time.Now()
	future := today.AddDate(0, 1, 1)

	certOld := &CertificateType{
		Dates: &CertificateDates{
			EndsAt: oldDate,
		},
	}
	err := checkExpiration(certOld)
	assert.Error(t, err, "error should not be nil")

	certToday := &CertificateType{
		Dates: &CertificateDates{
			EndsAt: today,
		},
	}

	err = checkExpiration(certToday)
	assert.Error(t, err, "error should not be nil")

	certFuture := &CertificateType{
		Dates: &CertificateDates{
			EndsAt: future,
		},
	}

	err = checkExpiration(certFuture)
	assert.NoError(t, err, "error should be nil")
}

func TestComparePublicKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating private key for testing : %v", err)
	}

	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating private key for testing : %v", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateKeyPEM2 := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey2),
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"My orga"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		fmt.Println("error while creating the certificate", err)
		return
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	err = comparePublicKeys(string(pem.EncodeToMemory(privateKeyPEM)), string(certPEM))
	assert.NoError(t, err, "err should be nil")

	err = comparePublicKeys(string(pem.EncodeToMemory(privateKeyPEM2)), string(certPEM))
	assert.Error(t, err, "err should not be nil")
}

func TestCompareCertificates(t *testing.T) {
	certificate1 := "    -----BEGIN CERTIFICATE-----\n    MIIG1jCCBT6gAwIBAgIRANKYpG4aqeIthQ5w7xVZzZIwDQYJKoZIhvcNAQEMBQAwVjELMAkGA1UE\n    BhMCRlIxDjAMBgNVBAoTBUdhbmRpMTcwNQYDVQQDEy5HYW5kaSBSU0EgRG9tYWluIFZhbGlkYXRp\n    b24gU2VjdXJlIFNlcnZlciBDQSAzMB4XDTIzMDkyNzAwMDAwMFoXDTI0MTAwOTIzNTk1OVowIjEg\n    MB4GA1UEAxMXd3d3LXN0YWdpbmcuZG9jdG9saWIuZGUwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\n    AASb0Mj9RTl3XXkCPG35HAklIXgRlFfK5EgH9YWQqGJs1UHz1ksujuzPiowQoRHzvvL+eVjkWnk7\n    LouugcX0UCcco4IEHDCCBBgwHwYDVR0jBBgwFoAUgRGS3mYypbBbMz1lQ4X81AQt8a4wHQYDVR0O\n    BBYEFLe+V4ib2FPbi/Aq3TwVLAL2rycXMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMB0G\n    A1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjBJBgNVHSAEQjBAMDQGCysGAQQBsjEBAgIaMCUw\n    IwYIKwYBBQUHAgEWF2h0dHBzOi8vc2VjdGlnby5jb20vQ1BTMAgGBmeBDAECATCBgwYIKwYBBQUH\n    AQEEdzB1ME4GCCsGAQUFBzAChkJodHRwOi8vY3J0LnNlY3RpZ28uY29tL0dhbmRpUlNBRG9tYWlu\n    VmFsaWRhdGlvblNlY3VyZVNlcnZlckNBMy5jcnQwIwYIKwYBBQUHMAGGF2h0dHA6Ly9vY3NwLnNl\n    Y3RpZ28uY29tMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkAdwB2/4g/Crb7lVHCYcz1h7o0tKTN\n    uyncaEIKn+ZnTFo6dAAAAYrXP3XLAAAEAwBIMEYCIQDCgbxW+ErC+4SFmrfUVO86DX/g6bjq7LqN\n    yd+Qb/vzwwIhAKcstQvNOcX8nC10dJKPuBQXhpSFtg+tcrqO2GwFl+SWAHYA2ra/az+1tiKfm8K7\n    XGvocJFxbLtRhIU0vaQ9MEjX+6sAAAGK1z92JwAABAMARzBFAiEA2KjugP4cLWJg4J+V/d5NKFWK\n    tnNZHzcqQEYrBNMT5T0CIB0K2NIM3mvZWnRZI9vToeJxdlR6QYg7U6VxUv762NWsAHYA7s3QZNXb\n    Gs7FXLedtM0TojKHRny87N7DUUhZRnEftZsAAAGK1z92QwAABAMARzBFAiBk74RSo3XNaeIk2LoF\n    QrsUxDCYvgELPJMpYkkMWmet3AIhAJEqr/osT0EqMR0YZsA4oLjyabpmNwwaid7vzGfqUCI9MIIB\n    QwYDVR0RBIIBOjCCATaCF3d3dy1zdGFnaW5nLmRvY3RvbGliLmRlghVhLXN0YWdpbmcuZG9jdG9s\n    aWIuZGWCGWFib3V0LXN0YWdpbmcuZG9jdG9saWIuZGWCF2FwaS1zdGFnaW5nLmRvY3RvbGliLmRl\n    ghphc3NldHMtc3RhZ2luZy5kb2N0b2xpYi5kZYIbY2FycmVycy1zdGFnaW5nLmRvY3RvbGliLmRl\n    ghhpbmZvLXN0YWdpbmcuZG9jdG9saWIuZGWCFW0tc3RhZ2luZy5kb2N0b2xpYi5kZYIccGFydG5l\n    cnMtc3RhZ2luZy5kb2N0b2xpYi5kZYIXcHJvLXN0YWdpbmcuZG9jdG9saWIuZGWCFnNiLXN0YWdp\n    bmcuZG9jdG9saWIuZGWCF3Rvay1zdGFnaW5nLmRvY3RvbGliLmRlMA0GCSqGSIb3DQEBDAUAA4IB\n    gQCDhYpSQyzWW2ZkkEpmczKZ3/w28jOarFkaEsWcXNQuY92I3kM83pE5hTOZuYjhoM964Lz0+7ad\n    CLz2+SvGwwDxNz7qvWGwcAElW6ALkhdZoLv0caC3DMdLxmTmUNiZspKjAOssC9mQIVAJa8q3/zbe\n    fSeuGdXCfPxbv+VSIGHW1/5/Ar47S+JRoNwjXGral4Crf8gWyYtYeeD+CEdIiw3wpRZ/REf2RGhT\n    /g0Os0PKDLxo/taZdXURvkoeny2OSHixcrurLSL+EUCtzWxNledIeSl1B4LYsBd++r7KikbOStY5\n    tyzje0GDAzEBvDtyWpWdpbC/tfNX36zRR7GhXBhv99jU+9C0uMW/6qoOfNOY0zbWMcP7n/xm6CEk\n    AvJASrnoSoip+AaH7tFgKAeACek5PHPlTr61FYch1PwbNm01lYgRynAc8sQcDA44SpmwPgyhpxYL\n    nlgOrSfdQ1oo+H8uRaqreOCV+L2GryRXa1W9CleuU9u2rY4l6ODOexWJ3Hs=\n    -----END CERTIFICATE-----\n    -----BEGIN CERTIFICATE-----\n    MIIGXDCCBESgAwIBAgIRAOkH5f+AdSJBCZB9ZyjKABAwDQYJKoZIhvcNAQEMBQAw\n    gYgxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpOZXcgSmVyc2V5MRQwEgYDVQQHEwtK\n    ZXJzZXkgQ2l0eTEeMBwGA1UEChMVVGhlIFVTRVJUUlVTVCBOZXR3b3JrMS4wLAYD\n    VQQDEyVVU0VSVHJ1c3QgUlNBIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MB4XDTIz\n    MDgwMjAwMDAwMFoXDTMzMDgwMTIzNTk1OVowVjELMAkGA1UEBhMCRlIxDjAMBgNV\n    BAoTBUdhbmRpMTcwNQYDVQQDEy5HYW5kaSBSU0EgRG9tYWluIFZhbGlkYXRpb24g\n    U2VjdXJlIFNlcnZlciBDQSAzMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKC\n    AYEAwrwuXKdKIiD9eu4fsNjLN0mS8HsTdDFyPPB5F5uUd6SJGutc7sqDd3T/p+gn\n    VoAZERvzAz8+OEux1GN1UJ+Gd8s5btXJCbDV5DpvzJOhfztk5JmFKz2XBka+MvDA\n    giyiZKs3G6yoMk8lEOu6NOsK3X8D1w0E6/C/ROa6Ml0ROnKm7vHGNVTfXTP5IqiN\n    h2JXmp4vD23gemf8nfuI2FngayMNsjm6SwpVYWfT3S8jn5el52FKzwo+uKVZAjNH\n    1ulgWoyO8p+PCsP+CvaEGDId3leSUVhPBBPRsxL42jjqo9aOKREgmrGco39JGf4O\n    ImxM8vKxQ9AjDrRTRETB9V9jbRf3v3Tojt3vBBwa3xQelVp9xUWQxo/5dV73g/c7\n    WWAvZ628XUw6k6vn6bY7qWuhehUO02plRLd5zP8nBORCbPmFCI97lZAnDYLprB4e\n    9IgCPJp+0zQDLr9o+eNKtR0a2Txb6nzGahIPi3a7QCH6+Yq4iwYVEQm+e6KBJZOm\n    +YiLAgMBAAGjggFwMIIBbDAfBgNVHSMEGDAWgBRTeb9aqitKz1SA4dibwJ3ysgNm\n    yzAdBgNVHQ4EFgQUgRGS3mYypbBbMz1lQ4X81AQt8a4wDgYDVR0PAQH/BAQDAgGG\n    MBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUF\n    BwMCMCIGA1UdIAQbMBkwDQYLKwYBBAGyMQECAhowCAYGZ4EMAQIBMFAGA1UdHwRJ\n    MEcwRaBDoEGGP2h0dHA6Ly9jcmwudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RSU0FD\n    ZXJ0aWZpY2F0aW9uQXV0aG9yaXR5LmNybDBxBggrBgEFBQcBAQRlMGMwOgYIKwYB\n    BQUHMAKGLmh0dHA6Ly9jcnQudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RSU0FBQUFD\n    QS5jcnQwJQYIKwYBBQUHMAGGGWh0dHA6Ly9vY3NwLnVzZXJ0cnVzdC5jb20wDQYJ\n    KoZIhvcNAQEMBQADggIBADvVncOMStREyA00ZSRUmrkmR3KzAlHVz06X1ydG9EpZ\n    z+JTQMWO809buLbDnr6t9z9jVnsDTQnWcMG4qiIkwhJVLxOVXUO+LFSBMskOe1SP\n    BtHwHS42DeZ8QTgbRlW9p/Ey9wIo+MS2tryQ9eaDTkc2FBed/82VjrdsQoeoTyuD\n    dp4tqarixjM/iJMgyEAMCpTkx4EqXJ/z7qgXusacsxMzt6NLv7FYcaKGbwjKqzrR\n    vEk/+ZYnZc5mxnautf0uwRCcOe0kCOh1fd+g6Tyd+cSj6oGcJY/f/Db0sxELpzGq\n    jRkbXan+eMojQfsgIe1n7SVyI5Yxz2RnQQL5ZT5K1mBcucqsTqkk3C7L3hF4hkwC\n    /Otm+badymHQcnbE1Pmz6ymqj2vtwT0mEQzetQdbvv3jc3ey4YcxirAM1ihxtXeI\n    NsEP1ndUV/0v+qqmk9iCoIjZQce8vAdziZqBYxO3NiZwTRAtqseiZWLJqQ077fy3\n    ebdjmw6y5U+DhDW2kxF/e+FJnu53DuY5/bE+oUneY770A7BfCuH+6uhEOaMNsn21\n    AHymLr1xlRPQYR0DMgHmsGTqdINcQfot1mlIXr05HQUK0b84CPgEU0zvVQL+j9dc\n    /4rh2sR6rl//tjG01Q+zQKStnR2NlNNrElDUC9IDmvL9JcF20cvOlE4R0lfTXa1k\n    -----END CERTIFICATE-----"
	certificate2 := " \t\t-----BEGIN CERTIFICATE-----\n    MIIG1jCCBT6gAwIBAgIRANKYpG4aqeIthQ5w7xVZzZIwDQYJKoZIhvcNAQEMBQAwVjELMAkGA1UE\n    BhMCRlIxDjAMBgNVBAoTBUdhbmRpMTcwNQYDVQQDEy5HYW5kaSBSU0EgRG9tYWluIFZhbGlkYXRp\n    b24gU2VjdXJlIFNlcnZlciBDQSAzMB4XDTIzMDkyNzAwMDAwMFoXDTI0MTAwOTIzNTk1OVowIjEg\n    MB4GA1UEAxMXd3d3LXN0YWdpbmcuZG9jdG9saWIuZGUwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\n    AASb0Mj9RTl3XXkCPG35HAklIXgRlFfK5EgH9YWQqGJs1UHz1ksujuzPiowQoRHzvvL+eVjkWnk7\n    LouugcX0UCcco4IEHDCCBBgwHwYDVR0jBBgwFoAUgRGS3mYypbBbMz1lQ4X81AQt8a4wHQYDVR0O\n    BBYEFLe+V4ib2FPbi/Aq3TwVLAL2rycXMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMB0G\n    A1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjBJBgNVHSAEQjBAMDQGCysGAQQBsjEBAgIaMCUw\n    IwYIKwYBBQUHAgEWF2h0dHBzOi8vc2VjdGlnby5jb20vQ1BTMAgGBmeBDAECATCBgwYIKwYBBQUH\n    AQEEdzB1ME4GCCsGAQUFBzAChkJodHRwOi8vY3J0LnNlY3RpZ28uY29tL0dhbmRpUlNBRG9tYWlu\n    VmFsaWRhdGlvblNlY3VyZVNlcnZlckNBMy5jcnQwIwYIKwYBBQUHMAGGF2h0dHA6Ly9vY3NwLnNl\n    Y3RpZ28uY29tMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkAdwB2/4g/Crb7lVHCYcz1h7o0tKTN\n    uyncaEIKn+ZnTFo6dAAAAYrXP3XLAAAEAwBIMEYCIQDCgbxW+ErC+4SFmrfUVO86DX/g6bjq7LqN\n    yd+Qb/vzwwIhAKcstQvNOcX8nC10dJKPuBQXhpSFtg+tcrqO2GwFl+SWAHYA2ra/az+1tiKfm8K7\n    XGvocJFxbLtRhIU0vaQ9MEjX+6sAAAGK1z92JwAABAMARzBFAiEA2KjugP4cLWJg4J+V/d5NKFWK\n    tnNZHzcqQEYrBNMT5T0CIB0K2NIM3mvZWnRZI9vToeJxdlR6QYg7U6VxUv762NWsAHYA7s3QZNXb\n    Gs7FXLedtM0TojKHRny87N7DUUhZRnEftZsAAAGK1z92QwAABAMARzBFAiBk74RSo3XNaeIk2LoF\n    QrsUxDCYvgELPJMpYkkMWmet3AIhAJEqr/osT0EqMR0YZsA4oLjyabpmNwwaid7vzGfqUCI9MIIB\n    QwYDVR0RBIIBOjCCATaCF3d3dy1zdGFnaW5nLmRvY3RvbGliLmRlghVhLXN0YWdpbmcuZG9jdG9s\n    aWIuZGWCGWFib3V0LXN0YWdpbmcuZG9jdG9saWIuZGWCF2FwaS1zdGFnaW5nLmRvY3RvbGliLmRl\n    ghphc3NldHMtc3RhZ2luZy5kb2N0b2xpYi5kZYIbY2FycmVycy1zdGFnaW5nLmRvY3RvbGliLmRl\n    ghhpbmZvLXN0YWdpbmcuZG9jdG9saWIuZGWCFW0tc3RhZ2luZy5kb2N0b2xpYi5kZYIccGFydG5l\n    cnMtc3RhZ2luZy5kb2N0b2xpYi5kZYIXcHJvLXN0YWdpbmcuZG9jdG9saWIuZGWCFnNiLXN0YWdp\n    bmcuZG9jdG9saWIuZGWCF3Rvay1zdGFnaW5nLmRvY3RvbGliLmRlMA0GCSqGSIb3DQEBDAUAA4IB\n    gQCDhYpSQyzWW2ZkkEpmczKZ3/w28jOarFkaEsWcXNQuY92I3kM83pE5hTOZuYjhoM964Lz0+7ad\n    CLz2+SvGwwDxNz7qvWGwcAElW6ALkhdZoLv0caC3DMdLxmTmUNiZspKjAOssC9mQIVAJa8q3/zbe\n    fSeuGdXCfPxbv+VSIGHW1/5/Ar47S+JRoNwjXGral4Crf8gWyYtYeeD+CEdIiw3wpRZ/REf2RGhT\n    /g0Os0PKDLxo/taZdXURvkoeny2OSHixcrurLSL+EUCtzWxNledIeSl1B4LYsBd++r7KikbOStY5\n    tyzje0GDAzEBvDtyWpWdpbC/tfNX36zRR7GhXBhv99jU+9C0uMW/6qoOfNOY0zbWMcP7n/xm6CEk\n    AvJASrnoSoip+AaH7tFgKAeACek5PHPlTr61FYch1PwbNm01lYgRynAc8sQcDA44SpmwPgyhpxYL\n    nlgOrSfdQ1oo+H8uRaqreOCV+L2GryRXa1W9CleuU9u2rY4l6ODOexWJ3Hs=\n    -----END CERTIFICATE-----\n    -----BEGIN CERTIFICATE-----\n    MIIGXDCCBESgAwIBAgIRAOkH5f+AdSJBCZB9ZyjKABAwDQYJKoZIhvcNAQEMBQAw\n    gYgxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpOZXcgSmVyc2V5MRQwEgYDVQQHEwtK\n    ZXJzZXkgQ2l0eTEeMBwGA1UEChMVVGhlIFVTRVJUUlVTVCBOZXR3b3JrMS4wLAYD\n    VQQDEyVVU0VSVHJ1c3QgUlNBIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MB4XDTIz\n    MDgwMjAwMDAwMFoXDTMzMDgwMTIzNTk1OVowVjELMAkGA1UEBhMCRlIxDjAMBgNV\n    BAoTBUdhbmRpMTcwNQYDVQQDEy5HYW5kaSBSU0EgRG9tYWluIFZhbGlkYXRpb24g\n    U2VjdXJlIFNlcnZlciBDQSAzMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKC\n    AYEAwrwuXKdKIiD9eu4fsNjLN0mS8HsTdDFyPPB5F5uUd6SJGutc7sqDd3T/p+gn\n    VoAZERvzAz8+OEux1GN1UJ+Gd8s5btXJCbDV5DpvzJOhfztk5JmFKz2XBka+MvDA\n    giyiZKs3G6yoMk8lEOu6NOsK3X8D1w0E6/C/ROa6Ml0ROnKm7vHGNVTfXTP5IqiN\n    h2JXmp4vD23gemf8nfuI2FngayMNsjm6SwpVYWfT3S8jn5el52FKzwo+uKVZAjNH\n    1ulgWoyO8p+PCsP+CvaEGDId3leSUVhPBBPRsxL42jjqo9aOKREgmrGco39JGf4O\n    ImxM8vKxQ9AjDrRTRETB9V9jbRf3v3Tojt3vBBwa3xQelVp9xUWQxo/5dV73g/c7\n    WWAvZ628XUw6k6vn6bY7qWuhehUO02plRLd5zP8nBORCbPmFCI97lZAnDYLprB4e\n    9IgCPJp+0zQDLr9o+eNKtR0a2Txb6nzGahIPi3a7QCH6+Yq4iwYVEQm+e6KBJZOm\n    +YiLAgMBAAGjggFwMIIBbDAfBgNVHSMEGDAWgBRTeb9aqitKz1SA4dibwJ3ysgNm\n    yzAdBgNVHQ4EFgQUgRGS3mYypbBbMz1lQ4X81AQt8a4wDgYDVR0PAQH/BAQDAgGG\n    MBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUF\n    BwMCMCIGA1UdIAQbMBkwDQYLKwYBBAGyMQECAhowCAYGZ4EMAQIBMFAGA1UdHwRJ\n    MEcwRaBDoEGGP2h0dHA6Ly9jcmwudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RSU0FD\n    ZXJ0aWZpY2F0aW9uQXV0aG9yaXR5LmNybDBxBggrBgEFBQcBAQRlMGMwOgYIKwYB\n    BQUHMAKGLmh0dHA6Ly9jcnQudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RSU0FBQUFD\n    QS5jcnQwJQYIKwYBBQUHMAGGGWh0dHA6Ly9vY3NwLnVzZXJ0cnVzdC5jb20wDQYJ\n    KoZIhvcNAQEMBQADggIBADvVncOMStREyA00ZSRUmrkmR3KzAlHVz06X1ydG9EpZ\n    z+JTQMWO809buLbDnr6t9z9jVnsDTQnWcMG4qiIkwhJVLxOVXUO+LFSBMskOe1SP\n    BtHwHS42DeZ8QTgbRlW9p/Ey9wIo+MS2tryQ9eaDTkc2FBed/82VjrdsQoeoTyuD\n    dp4tqarixjM/iJMgyEAMCpTkx4EqXJ/z7qgXusacsxMzt6NLv7FYcaKGbwjKqzrR\n    vEk/+ZYnZc5mxnautf0uwRCcOe0kCOh1fd+g6Tyd+cSj6oGcJY/f/Db0sxELpzGq\n    jRkbXan+eMojQfsgIe1n7SVyI5Yxz2RnQQL5ZT5K1mBcucqsTqkk3C7L3hF4hkwC\n    /Otm+badymHQcnbE1Pmz6ymqj2vtwT0mEQzetQdbvv3jc3ey4YcxirAM1ihxtXeI\n    NsEP1ndUV/0v+qqmk9iCoIjZQce8vAdziZqBYxO3NiZwTRAtqseiZWLJqQ077fy3\n    ebdjmw6y5U+DhDW2kxF/e+FJnu53DuY5/bE+oUneY770A7BfCuH+6uhEOaMNsn21\n    AHymLr1xlRPQYR0DMgHmsGTqdINcQfot1mlIXr05HQUK0b84CPgEU0zvVQL+j9dc\n /4rh2sR6rl//tjG01Q+zQKStnR2NlNNrElDUC9IDmvL9JcF20cvOlE4R0lfTXa1k\n -----END CERTIFICATE-----"
	certificate3 := "    -----BEGIN CERTIFICATE-----\n    AMIIG1jCCBT6gAwIBAgIRANKYpG4aqeIthQ5w7xVZzZIwDQYJKoZIhvcNAQEMBQAwVjELMAkGA1UE\n    BhMCRlIxDjAMBgNVBAoTBUdhbmRpMTcwNQYDVQQDEy5HYW5kaSBSU0EgRG9tYWluIFZhbGlkYXRp\n    b24gU2VjdXJlIFNlcnZlciBDQSAzMB4XDTIzMDkyNzAwMDAwMFoXDTI0MTAwOTIzNTk1OVowIjEg\n    MB4GA1UEAxMXd3d3LXN0YWdpbmcuZG9jdG9saWIuZGUwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\n    AASb0Mj9RTl3XXkCPG35HAklIXgRlFfK5EgH9YWQqGJs1UHz1ksujuzPiowQoRHzvvL+eVjkWnk7\n    LouugcX0UCcco4IEHDCCBBgwHwYDVR0jBBgwFoAUgRGS3mYypbBbMz1lQ4X81AQt8a4wHQYDVR0O\n    BBYEFLe+V4ib2FPbi/Aq3TwVLAL2rycXMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMB0G\n    A1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjBJBgNVHSAEQjBAMDQGCysGAQQBsjEBAgIaMCUw\n    IwYIKwYBBQUHAgEWF2h0dHBzOi8vc2VjdGlnby5jb20vQ1BTMAgGBmeBDAECATCBgwYIKwYBBQUH\n    AQEEdzB1ME4GCCsGAQUFBzAChkJodHRwOi8vY3J0LnNlY3RpZ28uY29tL0dhbmRpUlNBRG9tYWlu\n    VmFsaWRhdGlvblNlY3VyZVNlcnZlckNBMy5jcnQwIwYIKwYBBQUHMAGGF2h0dHA6Ly9vY3NwLnNl\n    Y3RpZ28uY29tMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkAdwB2/4g/Crb7lVHCYcz1h7o0tKTN\n    uyncaEIKn+ZnTFo6dAAAAYrXP3XLAAAEAwBIMEYCIQDCgbxW+ErC+4SFmrfUVO86DX/g6bjq7LqN\n    yd+Qb/vzwwIhAKcstQvNOcX8nC10dJKPuBQXhpSFtg+tcrqO2GwFl+SWAHYA2ra/az+1tiKfm8K7\n    XGvocJFxbLtRhIU0vaQ9MEjX+6sAAAGK1z92JwAABAMARzBFAiEA2KjugP4cLWJg4J+V/d5NKFWK\n    tnNZHzcqQEYrBNMT5T0CIB0K2NIM3mvZWnRZI9vToeJxdlR6QYg7U6VxUv762NWsAHYA7s3QZNXb\n    Gs7FXLedtM0TojKHRny87N7DUUhZRnEftZsAAAGK1z92QwAABAMARzBFAiBk74RSo3XNaeIk2LoF\n    QrsUxDCYvgELPJMpYkkMWmet3AIhAJEqr/osT0EqMR0YZsA4oLjyabpmNwwaid7vzGfqUCI9MIIB\n    QwYDVR0RBIIBOjCCATaCF3d3dy1zdGFnaW5nLmRvY3RvbGliLmRlghVhLXN0YWdpbmcuZG9jdG9s\n    aWIuZGWCGWFib3V0LXN0YWdpbmcuZG9jdG9saWIuZGWCF2FwaS1zdGFnaW5nLmRvY3RvbGliLmRl\n    ghphc3NldHMtc3RhZ2luZy5kb2N0b2xpYi5kZYIbY2FycmVycy1zdGFnaW5nLmRvY3RvbGliLmRl\n    ghhpbmZvLXN0YWdpbmcuZG9jdG9saWIuZGWCFW0tc3RhZ2luZy5kb2N0b2xpYi5kZYIccGFydG5l\n    cnMtc3RhZ2luZy5kb2N0b2xpYi5kZYIXcHJvLXN0YWdpbmcuZG9jdG9saWIuZGWCFnNiLXN0YWdp\n    bmcuZG9jdG9saWIuZGWCF3Rvay1zdGFnaW5nLmRvY3RvbGliLmRlMA0GCSqGSIb3DQEBDAUAA4IB\n    gQCDhYpSQyzWW2ZkkEpmczKZ3/w28jOarFkaEsWcXNQuY92I3kM83pE5hTOZuYjhoM964Lz0+7ad\n    CLz2+SvGwwDxNz7qvWGwcAElW6ALkhdZoLv0caC3DMdLxmTmUNiZspKjAOssC9mQIVAJa8q3/zbe\n    fSeuGdXCfPxbv+VSIGHW1/5/Ar47S+JRoNwjXGral4Crf8gWyYtYeeD+CEdIiw3wpRZ/REf2RGhT\n    /g0Os0PKDLxo/taZdXURvkoeny2OSHixcrurLSL+EUCtzWxNledIeSl1B4LYsBd++r7KikbOStY5\n    tyzje0GDAzEBvDtyWpWdpbC/tfNX36zRR7GhXBhv99jU+9C0uMW/6qoOfNOY0zbWMcP7n/xm6CEk\n    AvJASrnoSoip+AaH7tFgKAeACek5PHPlTr61FYch1PwbNm01lYgRynAc8sQcDA44SpmwPgyhpxYL\n    nlgOrSfdQ1oo+H8uRaqreOCV+L2GryRXa1W9CleuU9u2rY4l6ODOexWJ3Hs=\n    -----END CERTIFICATE-----\n    -----BEGIN CERTIFICATE-----\n    MIIGXDCCBESgAwIBAgIRAOkH5f+AdSJBCZB9ZyjKABAwDQYJKoZIhvcNAQEMBQAw\n    gYgxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpOZXcgSmVyc2V5MRQwEgYDVQQHEwtK\n    ZXJzZXkgQ2l0eTEeMBwGA1UEChMVVGhlIFVTRVJUUlVTVCBOZXR3b3JrMS4wLAYD\n    VQQDEyVVU0VSVHJ1c3QgUlNBIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MB4XDTIz\n    MDgwMjAwMDAwMFoXDTMzMDgwMTIzNTk1OVowVjELMAkGA1UEBhMCRlIxDjAMBgNV\n    BAoTBUdhbmRpMTcwNQYDVQQDEy5HYW5kaSBSU0EgRG9tYWluIFZhbGlkYXRpb24g\n    U2VjdXJlIFNlcnZlciBDQSAzMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKC\n    AYEAwrwuXKdKIiD9eu4fsNjLN0mS8HsTdDFyPPB5F5uUd6SJGutc7sqDd3T/p+gn\n    VoAZERvzAz8+OEux1GN1UJ+Gd8s5btXJCbDV5DpvzJOhfztk5JmFKz2XBka+MvDA\n    giyiZKs3G6yoMk8lEOu6NOsK3X8D1w0E6/C/ROa6Ml0ROnKm7vHGNVTfXTP5IqiN\n    h2JXmp4vD23gemf8nfuI2FngayMNsjm6SwpVYWfT3S8jn5el52FKzwo+uKVZAjNH\n    1ulgWoyO8p+PCsP+CvaEGDId3leSUVhPBBPRsxL42jjqo9aOKREgmrGco39JGf4O\n    ImxM8vKxQ9AjDrRTRETB9V9jbRf3v3Tojt3vBBwa3xQelVp9xUWQxo/5dV73g/c7\n    WWAvZ628XUw6k6vn6bY7qWuhehUO02plRLd5zP8nBORCbPmFCI97lZAnDYLprB4e\n    9IgCPJp+0zQDLr9o+eNKtR0a2Txb6nzGahIPi3a7QCH6+Yq4iwYVEQm+e6KBJZOm\n    +YiLAgMBAAGjggFwMIIBbDAfBgNVHSMEGDAWgBRTeb9aqitKz1SA4dibwJ3ysgNm\n    yzAdBgNVHQ4EFgQUgRGS3mYypbBbMz1lQ4X81AQt8a4wDgYDVR0PAQH/BAQDAgGG\n    MBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUF\n    BwMCMCIGA1UdIAQbMBkwDQYLKwYBBAGyMQECAhowCAYGZ4EMAQIBMFAGA1UdHwRJ\n    MEcwRaBDoEGGP2h0dHA6Ly9jcmwudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RSU0FD\n    ZXJ0aWZpY2F0aW9uQXV0aG9yaXR5LmNybDBxBggrBgEFBQcBAQRlMGMwOgYIKwYB\n    BQUHMAKGLmh0dHA6Ly9jcnQudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RSU0FBQUFD\n    QS5jcnQwJQYIKwYBBQUHMAGGGWh0dHA6Ly9vY3NwLnVzZXJ0cnVzdC5jb20wDQYJ\n    KoZIhvcNAQEMBQADggIBADvVncOMStREyA00ZSRUmrkmR3KzAlHVz06X1ydG9EpZ\n    z+JTQMWO809buLbDnr6t9z9jVnsDTQnWcMG4qiIkwhJVLxOVXUO+LFSBMskOe1SP\n    BtHwHS42DeZ8QTgbRlW9p/Ey9wIo+MS2tryQ9eaDTkc2FBed/82VjrdsQoeoTyuD\n    dp4tqarixjM/iJMgyEAMCpTkx4EqXJ/z7qgXusacsxMzt6NLv7FYcaKGbwjKqzrR\n    vEk/+ZYnZc5mxnautf0uwRCcOe0kCOh1fd+g6Tyd+cSj6oGcJY/f/Db0sxELpzGq\n    jRkbXan+eMojQfsgIe1n7SVyI5Yxz2RnQQL5ZT5K1mBcucqsTqkk3C7L3hF4hkwC\n    /Otm+badymHQcnbE1Pmz6ymqj2vtwT0mEQzetQdbvv3jc3ey4YcxirAM1ihxtXeI\n    NsEP1ndUV/0v+qqmk9iCoIjZQce8vAdziZqBYxO3NiZwTRAtqseiZWLJqQ077fy3\n    ebdjmw6y5U+DhDW2kxF/e+FJnu53DuY5/bE+oUneY770A7BfCuH+6uhEOaMNsn21\n    AHymLr1xlRPQYR0DMgHmsGTqdINcQfot1mlIXr05HQUK0b84CPgEU0zvVQL+j9dc\n    /4rh2sR6rl//tjG01Q+zQKStnR2NlNNrElDUC9IDmvL9JcF20cvOlE4R0lfTXa1k\n    -----END CERTIFICATE-----"

	assert.Equal(t, compareCertificates(certificate1, certificate2), false)
	assert.Equal(t, compareCertificates(certificate1, certificate3), true)
}
