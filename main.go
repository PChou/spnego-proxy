package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"gopkg.in/jcmturner/gokrb5.v7/client"
	"gopkg.in/jcmturner/gokrb5.v7/config"
	"gopkg.in/jcmturner/gokrb5.v7/keytab"
	"gopkg.in/jcmturner/gokrb5.v7/spnego"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func NewSingleHostReverseProxy(cl *client.Client, spn string, target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		err := spnego.SetSPNEGOHeader(cl, req, spn)
		if err != nil {
			log.Println(err)
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func main() {
	if len(os.Args) < 2 {
		panic("Unspecified config file as argument")
	}
	c := loadConfig(os.Args[1])
	c.checkValid()
	cfg, err := config.Load(c.Krb5)
	if err != nil {
		log.Printf("Failed to load krb5 config file: %v\n", err)
		os.Exit(1)
	}
	kt, err := keytab.Load(c.Client.Keytab)
	if err != nil {
		log.Printf("Failed to load keytab file: %v\n", err)
		os.Exit(1)
	}

	userName := ""
	realm := ""
	principalPart := strings.Split(c.Client.Principal, "@")
	if len(principalPart) > 1 {
		userName = principalPart[0]
		realm = principalPart[1]
	} else {
		userName = principalPart[0]
		realm = cfg.LibDefaults.DefaultRealm
	}

	cl := client.NewClientWithKeytab(userName, realm, kt, cfg)
	err = cl.Login()
	if err != nil {
		log.Printf("Failed to login: %v\n", err)
		os.Exit(1)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	remote, err := url.Parse(c.Server.Upstream)
	proxy := NewSingleHostReverseProxy(cl, c.Server.Principal, remote)
	http.Handle("/", &ProxyHandler{proxy})
	err = http.ListenAndServe(c.Server.Listen, nil)
	if err != nil {
		panic(err)
	}
}


type ProxyHandler struct {
	p *httputil.ReverseProxy
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	ph.p.ServeHTTP(w, r)
}