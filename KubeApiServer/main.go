package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/iamrz1/MockExtensionApiServer/Helpers"
	"io"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
	"time"
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
	//-----------------------------REQUEST----HEADER------------------------------------------
	//Request Header is invoked every time kubeApiServer wants to send a request through to DatabaseServer

	//----------------------------------SERVER-------------------------------------------------
	server,_:= Helpers.InitServer(IP,certificateStore.CertFile("ca"),
		certificateStore.CertFile("tls"),certificateStore.KeyFile("tls"))
	router := mux.NewRouter()
	router.HandleFunc("/", handler)
	router.HandleFunc("/core/{resource}", resourceHandler)
	router.HandleFunc("/db/{resource}", databaseHandler)
	server.ListenAndServe(router)

}

//----------------------HANDLER----FUNCTIONS--------------------------------------------------
func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(w, "OK")
	log.Println(w, "Recieved request")
}
func resourceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	log.Println(w, "Resource: ", vars["resource"])

}
func databaseHandler(w http.ResponseWriter, r *http.Request) {
	//Load Request Header TLS Certificate
	rhCertificate := GetRequestHeaderTlsCert()
	//Load CA of DataBaseServer from Files
	serverCACertPool := LoadServerCA()
	tr := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{rhCertificate},
			RootCAs:      serverCACertPool,
		},
	}
	//
	client := http.Client{
		Transport: tr,
		Timeout:   time.Duration(30 * time.Second),
	}

	u := *r.URL
	u.Scheme = "https"
	u.Host = "127.0.0.2:8443"

	fmt.Println("forwarding request to = ",u.String())

	req, _ := http.NewRequest(r.Method, u.String(), nil)
	if len(r.TLS.PeerCertificates) > 0 {
//		req.Header.Set("X-Remote-User", r.TLS.PeerCertificates[0].Subject.CommonName)
		fmt.Println("req.Header.Set was called")
		req.Header.Set("X-Remote-User", "Rezoan")
		req.Header.Set("u","name")
	}else{
		fmt.Println("req.Header.Set was not called. Peer cert len = ",len(r.TLS.PeerCertificates))

		req.Header.Set("X-Remote-User", "KubeApiServer")
	}

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error: %v\n", err.Error())
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusOK)
	io.Copy(w, resp.Body)
}
//router.HandleFunc("/core/{resource}", func(response http.ResponseWriter, request *http.Request) {
//	vars := mux.Vars(request)
//	response.WriteHeader(http.StatusOK)
//	fmt.Fprintf(response, "Resource: %v\n", vars["resource"])
//})

//------------------------------INIT-REQUEST-HEADER------------------------------------------
//GetRequestHeaderTlsCert returns a tls keyCertPair
func GetRequestHeaderTlsCert() tls.Certificate{

	rhCertStore,err := Helpers.NewCertStore("tmp/certificates/")
	err = rhCertStore.InitCA("rqheader","127.0.0.1")	//ip doesnt really matter.
	if err != nil {
		log.Println("Error creating request header CA")
		log.Fatalln(err)
	}
	rhClientCert, rhClientKey, err := rhCertStore.NewKeyCertPair(Helpers.ClientCert,cert.AltNames{
		DNSNames: []string{"kubeApiServer"}, // because rh will make calls on behalf of kubeApiServer
	})
	err = rhCertStore.Write("kubeApiServer", rhClientCert, rhClientKey)
	rhCert, err := tls.LoadX509KeyPair(rhCertStore.CertFile("kubeApiServer"), rhCertStore.KeyFile("kubeApiServer"))
	if err != nil {
		log.Println("Error creating request header tls certificate")
		log.Fatalln(err)
	}
	return rhCert
	//Note: DatabaseServer needs to have the request header CA we've just created (rqheader-ca)
	// in order to verify any request that is routed via (databaseHandler) using request header
	//
}
//-------------------------------LOAD-CA-FROM-DB-SERVER------------------------------------------
func LoadServerCA()*x509.CertPool{
	targetServerCACertPool :=x509.NewCertPool()
	easStore, err := Helpers.NewCertStore("tmp/certificates/")
	if err != nil {
		log.Fatalln(err)
	}
	//err = easStore.LoadCAForPrefix("DatabaseServer","127.0.0.2")
	err = easStore.InitCA("DatabaseServer","127.0.0.2")
	if err != nil {
		log.Fatalln(err)
	}
	targetServerCACertPool.AppendCertsFromPEM(easStore.CACertBytes())
	return targetServerCACertPool
}