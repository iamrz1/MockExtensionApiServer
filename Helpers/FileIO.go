package Helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"k8s.io/client-go/util/cert"
	"log"
)

func (s *CertStore) Read(name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	crtBytes, err := afero.ReadFile(s.fileSystem, s.CertFile(name))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to read certificate `%s`", s.CertFile(name))
	}
	log.Println("Reading certificate")
	crt, err := cert.ParseCertsPEM(crtBytes)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse certificate `%s`", s.CertFile(name))
	}

	keyBytes, err := afero.ReadFile(s.fileSystem, s.KeyFile(name))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to read private key `%s`", s.KeyFile(name))
	}
	key, err := cert.ParsePrivateKeyPEM(keyBytes)
	log.Println("Reading Key")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse private key `%s`", s.KeyFile(name))
	}
	return crt[0], key.(*rsa.PrivateKey), nil
}
//Writes CA specific crt and key in prefix-ca.crt file
//Write(s.ca="ca", crt=Certificate, key = PrivateKey)
func (s *CertStore) Write(name string, crt *x509.Certificate, key *rsa.PrivateKey) error {
	if err := afero.WriteFile(s.fileSystem, s.CertFile(name), cert.EncodeCertPEM(crt), 0644); err != nil {
		return errors.Wrapf(err, "failed to write cert `%s`", s.CertFile(name))
	}
	if err := afero.WriteFile(s.fileSystem, s.KeyFile(name), cert.EncodePrivateKeyPEM(key), 0600); err != nil {
		return errors.Wrapf(err, "failed to write key `%s`", s.KeyFile(name))
	}
	return nil
}