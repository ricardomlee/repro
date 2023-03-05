package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Proxies map[string]string
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.Host // 使用请求的域名作为反向代理的名称
		if proxy, ok := reverseProxies[name]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			http.Error(w, "No such proxy", http.StatusNotFound)
		}
	})

	// 启动HTTP服务器
	log.Println("Starting reverse proxy server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
