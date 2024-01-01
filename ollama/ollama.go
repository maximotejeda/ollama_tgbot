package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/maximotejeda/ollama_tgbot/dbx"
)

const (
	METHOD        = "POST"
	ADDRESS       = "http://localhost:11434"
	GENERATE_API  = "/api/generate"
	CHAT_API      = "/api/chat"
	DEFAULT_MODEL = "llama2-uncensored:7b-chat-q3_K_M"
)

var (
	ollamaUri             string = os.Getenv("OLLAMA_URI")
	ollamaPort            string = os.Getenv("OLLAMA_PORT")
	ollamaGeneratePath    string = os.Getenv("OLLAMA_GENERATE_PATH")
	ollamaChatPath        string = os.Getenv("OLLAMA_CHAT_PATH")
	ollamaChatDefaultTime string = os.Getenv("OLLAMA_DEFAULT_CHAT_TIME")
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Image   []byte `json:"images,omitempty"`
}
type parameters struct {
	Model    string    `json:"model"`
	Prompt   string    `json:"prompt,omitempty"`
	Format   string    `json:"format,omitempty"`
	Stream   bool      `json:"stream"`
	Messages []message `json:"messages,omitempty"`
	Images   []byte    `json:"images,omitempty"`
	Raw      bool      `json:"raw,omitempty"`
}

type ollamaClient struct {
	ctx  context.Context
	db   *dbx.DB
	log  *slog.Logger
	user *dbx.User
}

type oResponse struct {
	Model              string   `json:"model"`
	CreatedAt          string   `json:"created_at"`
	Message            *message `json:"message,omitempty"`
	Response           string   `json:"response,omitempty"`
	Done               bool     `json:"done"`
	Context            []int    `json:"context,omitempty"`
	TotalDuration      int64    `json:"total_duration"`
	LoadDuration       int64    `json:"load_duration"`
	PromptEvalCount    int      `json:"prompt_eval_count"`
	PromptEvalDuration int64    `json:"prompt_eval_duration"`
	EvalCount          int      `json:"eval_count"`
	EvalDuration       int64    `json:"eval_duration"`
}

func NewOllamaClient(ctx context.Context, db *dbx.DB, log *slog.Logger, user *dbx.User) *ollamaClient {
	return &ollamaClient{ctx: ctx, db: db, log: log, user: user}
}

func (o *ollamaClient) Do(content string) string {
	// create uri
	baseUri := ollamaUri + ":" + ollamaPort
	o.log.Info("query selected", "type", o.user.Mode)
	if o.user.Mode == "chat" {
		baseUri = baseUri + ollamaChatPath
	} else {
		baseUri = baseUri + ollamaGeneratePath
		param := parameters{
			Model:  o.user.Model,
			Prompt: content,
			Stream: false,
		}
		bodyBytes, err := json.Marshal(param)
		body := io.NopCloser(bytes.NewReader(bodyBytes))
		if err != nil {
			o.log.Error("[client.DO] marshaling struct", "error", err.Error())
			return ""
		}
		str := OllamaGenerate(baseUri, body)
		return str

	}

	// get info needed to query
	// TODO: Limit the amount of time to look by env
	h := dbx.NewHistory(o.ctx, o.db, o.log)
	hList, err := h.Query(*o.user)
	messageList := []message{}
	if err != nil {
		o.log.Error("[client.DO]error on query history", "error", err.Error())
	}

	for _, h := range hList {
		m := message{Role: h.Role, Content: h.Conversation}
		messageList = append(messageList, m)
	}
	m := message{Role: "user", Content: content}
	messageList = append(messageList, m)
	o.log.Info("chat Length", "length", len(messageList))
	param := parameters{
		Model:    o.user.Model,
		Messages: messageList,
		Stream:   false,
	}

	bodyBytes, err := json.Marshal(param)
	o.log.Warn("Request to chat ollama", "value", string(bodyBytes))
	body := io.NopCloser(bytes.NewReader(bodyBytes))
	if err != nil {
		o.log.Error("[client.DO] marshaling struct", "error", err.Error())
		return ""
	}

	str := chat(baseUri, body)
	strCopy := strings.Clone(str)
	if strings.ReplaceAll(strCopy, "\n", "") != "" {
		defer h.Add(*o.user, strCopy, "assistant")
		defer h.Add(*o.user, content, "user")
	} else {
		str = "empty ollama response, try other input if this keep ocurring try /reset the context"
	}
	return str
}

func OllamaGenerate(uri string, body io.Reader) string {
	req, err := http.NewRequest(METHOD, uri, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	if res.StatusCode != 200 {
		fmt.Println("error", res.StatusCode)
	} else {
		defer res.Body.Close()
		bt, _ := io.ReadAll(res.Body)
		resp := oResponse{}
		err := json.Unmarshal(bt, &resp)
		if err != nil {
			log.Println(err, string(bt))
		}

		return resp.Response
	}
	return ""
}

// chat:
// will work as a chat to mantain the context over the conversation
func chat(uri string, body io.Reader) string {
	req, err := http.NewRequest(METHOD, uri, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		//log.Fatal("algo estuvo mal", err.Error())
		return "algo fallo"
	}

	if res.StatusCode != 200 {
		return "code different to 20"
		fmt.Println("error status not 200", res.StatusCode)
	} else {
		defer res.Body.Close()
		resp := oResponse{}

		bt, _ := io.ReadAll(res.Body)
		err := json.Unmarshal(bt, &resp)
		if err != nil {
			log.Println(err, string(bt))
		}

		log.Println("empty response", "response", fmt.Sprintf("%+v", resp))
		return resp.Message.Content
	}
	return "something went wrong"
}
