package chat

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/sashabaranov/go-openai"

	"github.com/lenye/aichat/internal/chatgpt"
	"github.com/lenye/aichat/internal/config"
	"github.com/lenye/aichat/pkg/web/logging"
	"github.com/lenye/aichat/pkg/web/render"
	"github.com/lenye/aichat/pkg/web/templatemap"
)

// Message 直接回复聊天响应消息 sse
func Message(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m := templatemap.FromContext(ctx)
	logger := logging.FromContext(ctx)

	flusher, err := w.(http.Flusher)
	if !err {
		logger.Error("streaming unsupported")
		render.Html500(w, r, errors.New("streaming unsupported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// nginx 添加X-Accel-Buffering=no的响应header，来告诉nginx不要对响应数据进行缓存。
	w.Header().Set("X-Accel-Buffering", "no")

	streamID := r.PostFormValue("stream_id")
	prompt := r.PostFormValue("prompt")
	if prompt != "" {
		in := &chatgpt.Message{
			StreamID:  streamID,
			ID:        "",
			User:      "",
			Model:     "",
			Prompt:    prompt,
			System:    "",
			Stream:    false,
			History:   0,
			MaxTokens: 0,
		}

		in.Stream, _ = strconv.ParseBool(r.PostFormValue("stream"))
		in.Model = r.PostFormValue("model")
		if in.Model == "" {
			in.Model = openai.GPT3Dot5Turbo
		}
		in.System = r.PostFormValue("system")
		// if uHis, err := strconv.ParseUint(r.PostFormValue("history"), 10, 0); err != nil {
		// 	in.History = uint(uHis)
		// }
		if uu, err := strconv.ParseUint(r.PostFormValue("max_tokens"), 10, 0); err == nil {
			in.MaxTokens = uint(uu)
		}

		logger.Debug("input",
			"data", in,
		)

		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		var hisMsg []openai.ChatCompletionMessage // 聊天记录
		chatReq := chatgpt.MakeChatRequest(in, hisMsg)
		chStr := make(chan string)

		var wg sync.WaitGroup
		wg.Add(1)
		// ai chat
		go func() {
			defer wg.Done()
			chatgpt.HttpChatCompletion(r, config.Default().OpenAI, chatReq, chStr)
		}()
		messages := chatgpt.HttpChatResponseProcess(w, r, chStr)
		logger.Debug("ai",
			"msg", messages,
		)

		wg.Wait()

		// todo 计算token，保存账户余额，保存聊天记录

		m["stream_id"] = in.StreamID
		m["model"] = in.Model
		m["stream"] = strconv.FormatBool(in.Stream)
		m["system"] = in.System
		m["history"] = strconv.FormatUint(uint64(in.History), 10)
	} else {
		m["stream_id"] = getStreamID(w, r)
		m["model"] = openai.GPT3Dot5Turbo
		m["stream"] = "true"
		m["system"] = ""
		m["history"] = "0"
	}

	render.Html(w, r, "chat_input.gohtml", m)
}
