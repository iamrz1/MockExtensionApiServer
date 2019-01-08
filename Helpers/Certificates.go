package Helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"os"
	"strings"
)
type CertStore struct {
	fileSystem           afero.Fs
	directory          string
	prefix       string
	ca           string
	caKey        *rsa.PrivateKey
	caCert       *x509.Certificate
}
type CertificateType string

const ClientCert CertificateType = "client"
const ServerCert CertificateType = "Server"

func NewCertStore( directory string) (*CertStore, error){
	//for the giver directory string, create a directory in the fileSystem
	fileSystem := afero.NewOsFs()
	err := fileSystem.MkdirAll(directory, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dir `%s`", directory)
	}
	return &CertStore{fileSystem: fileSystem, directory: directory, ca: "ca", }, nil
}

func (s *CertStore) InitCA(prefix string, ip string) error {
	if len(prefix)>0{
		s.prefix=strings.ToLower(strings.TrimSpace(prefix)) + "-"
	}else{
		return  errors.Errorf("Insert a valid Prefix of length>0")
	}
	err := s.LoadCA()
	if err == nil {
		log.Println("CA Loaded")
		return nil
	}
	log.Println("Creating New CA")
	return s.NewCA(ip)
}
func (s *CertStore) LoadCA() error {

	if s.KeyCertPairExists(s.ca) {
		var err error
		s.caCert, s.caKey, err = s.Read(s.ca)
		return err
	}
	// Pair doesnt exist
	return os.ErrNotExist
}
func (s *CertStore) NewCA(ip string) error {
	key, err := cert.NewPrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate private key")
	}
	return s.createCAFromKey(key,ip)
}

func (s *CertStore) createCAFromKey(key *rsa.PrivateKey,ip string) error {
	var err error
	caCertConfig := s.getCACertificateConfig(ip)
	crt, err := cert.NewSelfSignedCACert(caCertConfig, key)
	if err != nil {
		return errors.Wrap(err, "failed to generate self-signed certificate")
	}
	//write ca crt to file
	err = s.Write(s.ca, crt, key)
	if err != nil {
		return err
	}

	s.caCert = crt
	s.caKey = key
	return nil
}

func (s *CertStore) getCACertificateConfig(ip string) cert.Config {

	cfg := cert.Config{
		CommonName:   s.ca,
		Organization: []string {"",},
		AltNames: cert.AltNames{
			DNSNames: []string{s.ca},
			IPs:      []net.IP{net.ParseIP(ip)},
		},
	}
	return cfg
}

func (s *CertStore) KeyCertPairExists(fileName string) bool {
	if _, err := s.fileSystem.Stat(s.CertFile(s.ca)); err == nil {
		if _, err := s.fileSystem.Stat(s.KeyFile(s.ca)); err == nil {
			return true
		}
	}
	return false
}

//NewServerCertPair(certificateType = "client/server", lookupName = DNS/IP address, lookUpType ="DNS/IP")
func (s *CertStore) NewKeyCertPair(certificateType CertificateType, lookUpNames cert.AltNames) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	certificateConfig := getCertificateConfig(certificateType,lookUpNames)
	crt, err := cert.NewSignedCert(certificateConfig, key, s.caCert, s.caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate server certificate")
	}
	return crt, key, nil
}
//getCertificateConfig returns configuration for a new certificate pair
func getCertificateConfig(certificateType CertificateType, lookUpNames cert.AltNames) cert.Config{
	var usage = x509.ExtKeyUsageServerAuth
	if certificateType=="client"{
		usage = x509.ExtKeyUsageClientAuth
	}
	var commonName string

	if len(lookUpNames.DNSNames) > 0 {
		commonName = lookUpNames.DNSNames[0]
		log.Print("Extracting common name = "+ commonName+" from DNS")
	}else if len(lookUpNames.IPs) > 0 {
		commonName = lookUpNames.IPs[0].String()
		log.Print("Extracting common name = "+ commonName+" from IP")
	}else{
		commonName= "undefinedCN"
	}

	cfg := cert.Config{
		CommonName:   commonName,
		AltNames:     lookUpNames,
		Usages:       []x509.ExtKeyUsage{usage},
	}
	log.Println("certificate Configured.")
	return cfg
}