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
	Cert    struct {
		Dir   string
		Cache string
	}
}

func main() {
	// 读取配置文件
	config := Config{}
	if _, err := toml.DecodeFile("config/repro.toml", &config); err != nil {
		log.Fatal(err)
	}

	var domains []string

	// 创建反向代理实例
	reverseProxies := make(map[string]*httputil.ReverseProxy)
	for name, targetURL := range config.Proxies {
		target, err := url.Parse(targetURL)
		if err != nil {
			log.Fatalf("Invalid target URL for %s: %s", name, err)
		}
		reverseProxies[name] = httputil.NewSingleHostReverseProxy(target)
		domains = append(domains, name)
		log.Println("Add proxy for", name)
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

	if config.Cert.Cache != "" { // 使用HTTPS, autocert生成证书
		log.Println("autocert cache dir", config.Cert.Cache)
		// 设置证书管理器
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(config.Cert.Cache),
			HostPolicy: autocert.HostWhitelist(domains...),
		}

		// 创建HTTPS服务器
		tlsConfig := &tls.Config{
			GetCertificate: m.GetCertificate,
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
			Handler:   m.HTTPHandler(nil),
			TLSConfig: tlsConfig,
		}
		log.Println("Starting HTTP server on :http")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	} else if config.Cert.Dir != "" { // 使用HTTPS, 自定义证书目录
		// 加载证书和私钥文件
		cert, err := tls.LoadX509KeyPair(
			config.Cert.Dir+"/cert.pem",
			config.Cert.Dir+"/key.pem",
		)
		if err != nil {
			log.Fatal("Failed to load certificate and key: ", err)
		}

		// 创建自定义的 TLS 配置
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
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
			Addr: ":http",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusPermanentRedirect)
			}),
		}
		log.Println("Starting HTTP server on :http")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	} else { // 使用HTTP
		// 创建HTTP服务器
		httpServer := &http.Server{
			Addr:    ":http",
			Handler: mux,
		}
		log.Println("Starting HTTP server on :http")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
}
