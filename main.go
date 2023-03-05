package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	Proxies map[string]string
	CertDir string
}

func main() {
	// 读取配置文件
	config := Config{}
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatal(err)
	}

	// 创建反向代理实例
	reverseProxies := make(map[string]*httputil.ReverseProxy)
	for name, targetURL := range config.Proxies {
		target, err := url.Parse(targetURL)
		if err != nil {
			log.Fatalf("Invalid target URL for %s: %s", name, err)
		}
		reverseProxies[name] = httputil.NewSingleHostReverseProxy(target)
	}

	// 创建HTTP服务器并注册反向代理处理器
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.Host // 使用请求的域名作为反向代理的名称
		if proxy, ok := reverseProxies[name]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			http.Error(w, "No such proxy", http.StatusNotFound)
		}
	})

	// 创建自动证书管理器
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(config.CertDir),
	}

	// 创建HTTPS服务器
	tlsConfig := &tls.Config{
		GetCertificate: certManager.GetCertificate,
	}
	server := &http.Server{
		Addr:      ":https",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// 启动HTTPS服务器
	go func() {
		log.Println("Starting HTTPS server on :https")
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	// 创建HTTP服务器并重定向到HTTPS服务器
	httpServer := &http.Server{
		Addr:      ":http",
		Handler:   certManager.HTTPHandler(nil),
		TLSConfig: &tls.Config{GetCertificate: certManager.GetCertificate},
	}
	log.Println("Starting HTTP server on :http")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
