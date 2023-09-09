# aichat

`aichat`是 OpenAI GPT API 驱动的人工智能聊天，`aichat`: 获得即时答案，寻找创意灵感，并学习新东西。

```shell
$ ./aichat -h

Usage:
  aichat [flags]

Flags:
  -h, --help                         help for aichat
      --log_format string            log message encode format: text, json (default "text")
      --log_level string             log message level: debug, info, warn, error (default "info")
      --mode string                  running mode: console, web (default "console")
      --openai_api_base_url string   openai api base url
      --openai_api_key string        openai api key (required)
      --openai_api_type string       openai api type: open_ai, azure (default "OPEN_AI")
      --openai_history uint          openai chat message history
      --openai_max_tokens uint       openai chat message max tokens
      --openai_model string          openai chat message model (default "gpt-3.5-turbo")
      --openai_proxy string          openai proxy
      --openai_stream                openai chat message stream mode (default true)
      --openai_system string         openai chat message system prompt
  -v, --version                      version for aichat
      --web_port uint                web server listen port (default 8080)
```

两种代理说明：

1. --openai_proxy 直接代理示例: http://127.0.0.1:9080 或者 socks5://127.0.0.1:1080
2. --openai_api_base_url 使用反向代理 https://github.com/lenye/chatgpt_reverse_proxy

### 命令行模式

```shell
./aichat --openai_api_key=xxx
---------------------
>
```

### web模式

```shell
./aichat --openai_api_key=xxx --mode=web
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
         - --mode=web   
         - --openai_api_key=XXX
   ```

## 源代码

```shell
$ git clone https://github.com/lenye/aichat.git
```

## License

`aichat` is released under the [Apache 2.0 license](https://github.com/lenye/aichat/blob/v0.3.0/LICENSE). 