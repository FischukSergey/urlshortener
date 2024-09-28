package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// Генерация сертификата и ключа
func GenerateTLS() error {
	//Генерация информации о сертификате
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2024),
		Subject: pkix.Name{
			Organization: []string{"yandexpracticum"},
			Country:      []string{"RU"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Генерация ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	//Генерация сертификата
	//cert, cert означает, что сертификат самоподписанный
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	//Генерация сертификата и ключа в PEM формате
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return fmt.Errorf("ошибка кодирования сертификата: %w", err)
	}
	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return fmt.Errorf("ошибка кодирования ключа: %w", err)
	}

	//Создание и запись файла сертификата
	certOut, err := os.Create("server.crt")
	if err != nil {
		return fmt.Errorf("ошибка создания файла server.crt: %w", err)
	}
	defer func() {
		if err := certOut.Close(); err != nil {
			fmt.Println("Ошибка закрытия файла server.crt: %w", err)
		}
	}()
	_, err = certOut.Write(certPEM.Bytes())
	if err != nil {
		return fmt.Errorf("ошибка записи в файл server.crt: %w", err)
	}

	//Создание и запись файла ключа
	keyOut, err := os.Create("server.key")
	if err != nil {
		return fmt.Errorf("ошибка создания файла server.key: %w", err)
	}
	defer func() {
		if err := keyOut.Close(); err != nil {
			fmt.Println("Ошибка закрытия файла server.key: %w", err)
		}
	}()
	_, err = keyOut.Write(privateKeyPEM.Bytes())
	if err != nil {
		return fmt.Errorf("ошибка записи в файл server.key: %w", err)
	}

	return nil
}
