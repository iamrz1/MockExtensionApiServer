package Helpers

import (
	"strings"
)

func (s *CertStore) CertFile(fileName string) string  {
	return s.directory+s.prefix + strings.ToLower(fileName) + ".crt"
}
func (s *CertStore) KeyFile(fileName string) string  {
	return s.directory+s.prefix + strings.ToLower(fileName) + ".key"
}
func InitServer(address string,ca string, tlsCert string, tlsKey string) *GenericServer{
	cfg := Config{
		Address: address+":8443",
		CACertFiles: []string{
			ca,
		},
		CertFile:tlsCert,
		KeyFile:  tlsKey,
	}
	srv := NewGenericServer(cfg)
	return srv
}