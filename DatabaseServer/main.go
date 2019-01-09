package main

import (
	"crypto/x509"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/iamrz1/MockExtensionApiServer/Helpers"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
)


func main(){
	const IP string = "127.0.0.2"
	certificateStore, err := Helpers.NewCertStore("tmp/certificates/")
	//Initialize CA
	err = certificateStore.InitCA("DatabaseServer",IP)
	if err != nil {
		log.Fatalln(err)
	}
	//Set TLS Server credentials using newly created CA
	tlsServerCert, tlsServerKey, err := certificateStore.NewKeyCertPair(Helpers.ServerCert, cert.AltNames{
		IPs: []net.IP{net.ParseIP(IP)},
	})
	err = certificateStore.Write("tls", tlsServerCert, tlsServerKey)
	if err != nil {
		log.Println("Error obtaining or writing TLS Server Certificates")
		log.Fatalln(err)
	}

	clientCert, clientKey, err := certificateStore.NewKeyCertPair(Helpers.ClientCert,cert.AltNames{
		DNSNames:[]	string{"xyz-client"},
	})
	err = certificateStore.Write("kubeApiServer", clientCert, clientKey)
	if err != nil {
		log.Fatalln(err)
	}


	cfg := Helpers.Config{
		Address: IP+":8443",
		CACertFiles: []string{
			certificateStore.CertFile("ca"),
		},
		CertFile:certificateStore.CertFile("tls"),
		KeyFile:  certificateStore.KeyFile("tls"),
	}

	cfg.CACertFiles = append(cfg.CACertFiles, certificateStore.CertFile("ca"))
	log.Println("CFG Before= ",cfg.CACertFiles)
	rhCACertPool := x509.NewCertPool()
	rhStore, err := Helpers.NewCertStore("tmp/certificates/")
	if err != nil {
		log.Fatalln(err)
	}

	err = rhStore.InitCA("rqheader","127.0.0.1")
	if err != nil {
		log.Fatalln(err)
	}else{
		log.Println("rq header CA loaded")
	}
	rhCACertPool.AppendCertsFromPEM(rhStore.CACertBytes())
	cfg.CACertFiles = append(cfg.CACertFiles, rhStore.CertFile("ca"))
	log.Println("CFG After= ",cfg.CACertFiles)
	server := Helpers.NewGenericServer(cfg)
	router := mux.NewRouter()
	router.HandleFunc("/", handler)
	router.HandleFunc("/db/{resource}", func (w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		w.WriteHeader(http.StatusOK)
		log.Println(w, "Resource: ", vars["resource"])

		user := "system:anonymous"
		src := "-"

		if len(r.TLS.PeerCertificates) > 0 { // client TLS was used
			opts := x509.VerifyOptions{
				Roots:     rhCACertPool,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}
			if _, err := r.TLS.PeerCertificates[0].Verify(opts); err != nil {
				user = r.TLS.PeerCertificates[0].Subject.CommonName // user name from client cert
				src = "Client-Cert-CN"
			} else {
				user = r.Header.Get("X-Remote-User") // user name from header value passed by apiserver
				src = "X-Remote-User"
			}
		}
		fmt.Fprintf(w, "Resource: %v requested by user[%s]=%s\n", vars["resource"], src, user)
		log.Println( "Resource: ", vars["resource"]," requested by source = ", src,"User = ", user)

	})
	server.ListenAndServe(router)

}
func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(w, "OK")
	log.Println(w, "Recieved request")
}


//router.HandleFunc("/core/{resource}", func(response http.ResponseWriter, request *http.Request) {
//	vars := mux.Vars(request)
//	response.WriteHeader(http.StatusOK)
//	fmt.Fprintf(response, "Resource: %v\n", vars["resource"])
