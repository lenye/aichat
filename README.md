# aichat

`aichat`是 OpenAI GPT API 驱动的人工智能聊天，`aichat`: 获得即时答案，寻找创意灵感，并学习新东西。

```shell
$ ./aichat -h

Usage:
  aichat [flags]

Flags:
      --console                      running console mode
  -h, --help                         help for aichat
      --httpd_port uint              httpd listen port (default 8080)
      --log_caller                   log annotate each message with the filename, line number and function name
      --log_format string            log message encode format: TEXT, JSON (default "TEXT")
      --log_level string             log message level: DEBUG, INFO, WARN, ERROR (default "INFO")
      --openai_api_base_url string   openai api base url
      --openai_api_key string        openai api key (required)
      --openai_api_type string       openai api type: OPEN_AI, AZURE (default "OPEN_AI")
      --openai_history uint          openai chat message history
      --openai_max_tokens uint       openai chat message max tokens
      --openai_model string          openai chat message model (default "gpt-3.5-turbo")
      --openai_proxy string          openai proxy
      --openai_stream                openai chat message stream mode (default true)
      --openai_system string         openai chat message role system
  -v, --version                      version for aichat
```

两种代理说明：

1. --openai_proxy 直接代理示例: http://127.0.0.1:9080 或者 socks5://127.0.0.1:1080
2. --openai_api_base_url 使用反向代理 https://github.com/lenye/chatgpt_reverse_proxy

### 命令行模式

```shell
./aichat --openai_api_key=xxx --console
---------------------
>
```

### web模式

```shell
./aichat --openai_api_key=xxx
time=2023-08-07T12:42:20.099+08:00 level=INFO msg="http server listening on [::]:8080"
```

浏览器访问: http://localhost:8080/chat

web不支持的参数: --openai_history

## Docker

1. 拉取容器映像
   ```shell
   $ docker pull ghcr.io/lenye/aichat
   ```

1. 开始运行它
   ```shell
   $ docker run --rm ghcr.io/lenye/aichat --help
   ```

1. docker-compose.yml
   ```yaml
   services:
   
     aichat:
       image: ghcr.io/lenye/aichat
       restart: unless-stopped
       ports:
         - "8080:8080"    
       volumes:
         - /etc/localtime:/etc/localtime:ro
       command:
         - --openai_api_key=XXX
   
   ```

## 源代码

```shell
$ git clone https://github.com/lenye/aichat.git
```

## License

`aichat` is released under the [Apache 2.0 license](https://github.com/lenye/aichat/blob/v0.3.0/LICENSE). 