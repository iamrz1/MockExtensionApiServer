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
