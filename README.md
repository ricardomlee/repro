# repro 反向代理服务器

![icon from midjourney](doc/img/repro.png)

这是一个基于Go语言实现的反向代理服务器，可以将多个域名的请求转发到不同的目标服务器上，以提高网站的性能、可靠性和安全性。

## 功能特点

- 支持多个反向代理目标，可以根据域名将请求转发到不同的目标服务器上。
- 支持负载均衡功能，可以将请求分发到多个目标服务器上，以提高性能和可靠性。
- 支持缓存功能，可以缓存目标服务器的响应，以减少响应时间和网络带宽的消耗。
- 支持安全功能，可以过滤和阻止一些恶意请求，例如SQL注入、XSS攻击、DDoS攻击等。
- 支持HTTPS功能，可以自动管理证书，实现自动化证书申请和更新。

## 使用方法

### 安装和配置

1. 下载源代码并解压缩到本地目录。

2. 安装Go语言并设置环境变量。

3. 打开命令行终端，进入源代码目录。

4. 执行以下命令，安装依赖库：

    ```shell
    go mod download
    ```

5. 修改配置文件config/repro.toml，配置反向代理目标和证书缓存目录。

6. 执行以下命令，启动反向代理服务器：

    ```shell
    go run main.go
    ```

### 配置文件说明

配置文件repro.toml使用TOML格式，包含以下配置项：

```toml
[proxies]
"example.com" = "http://localhost:8080"
"api.example.com" = "http://localhost:8081"

[cert]
dir = "config/cert"
```

其中，`[proxies]`部分定义了反向代理目标，使用域名作为键名，URL作为键值。`[cert]`部分定义了证书缓存目录。

### HTTPS配置

如果需要启用HTTPS功能，可以在配置文件中添加以下配置项：

```toml
[cert]
dir = "/config/cert"
```

其中，`dir`键指定证书缓存目录。反向代理服务器会自动管理证书，实现自动化证书申请和更新

### 容器化部署

如果需要使用容器化 (token用完了...)

要构建Docker镜像，可以使用以下命令：

```shell
docker build -t repro .
```

其中，-t参数用于指定镜像名称和标签。在这个示例中，我们将镜像命名为repro。

要运行Docker容器，可以使用以下命令：

```shell
docker run -p 80:80 -p 443:443 -v /mnt/user/appdata/repro:/app/config repro
```

其中，`-p`参数用于指定容器端口和主机端口的映射关系，`-v`参数用于指定容器目录和主机目录的映射关系。在这个示例中，我们将容器的80端口和443端口映射到主机的80端口和443端口，将含有repro.toml的配置文件目录映射到容器的`/app/config`目录。

需要注意的是，如果使用自动证书管理器，需要确保证书缓存目录`config/cert`具有足够的读写权限，以便自动证书管理器可以保存证书。

# Acknowledgements

Original code generated by [ChatGPT](https://chat.openai.com/)

Icon by [Midjourney](https://midjourney.com/)
