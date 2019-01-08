package main

import (
	"github.com/gorilla/mux"
	"github.com/iamrz1/MockExtensionApiServer/Helpers"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
)

func main(){
	const IP string = "127.0.0.1"
	certificateStore, err := Helpers.NewCertStore("tmp/certificates/")
	//Initialize CA
	err = certificateStore.InitCA("kubeApiServer",IP)
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
		DNSNames:[]	string{"rezoan"},
	})
	err = certificateStore.Write("rezoan", clientCert, clientKey)
	if err != nil {
		log.Fatalln(err)
	}
	server:= Helpers.InitServer(IP,certificateStore.CertFile("ca"),
		certificateStore.CertFile("tls"),certificateStore.KeyFile("tls"))

	router := mux.NewRouter()
	router.HandleFunc("/", handler)
	router.HandleFunc("/core/{resource}", resourceHandler)
	server.ListenAndServe(router)

}
func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(w, "OK")
	log.Println(w, "Recieved request")
}
func resourceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	log.Println(w, "Resource: ", vars["resource"])

}
//router.HandleFunc("/core/{resource}", func(response http.ResponseWriter, request *http.Request) {
//	vars := mux.Vars(request)
//	response.WriteHeader(http.StatusOK)
//	fmt.Fprintf(response, "Resource: %v\n", vars["resource"])
//})