package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	METHOD        = "POST"
	ADDRESS       = "http://localhost:11434"
	PATH          = "/api/generate"
	DEFAULT_MODEL = "llama2-uncensored:7b-chat-q3_K_M"
)

func OllamaConsult(ctx context.Context, prompt string) string {
	req, err := http.NewRequest(METHOD, ADDRESS+PATH, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.Clone(ctx)
	data := map[string]interface{}{
		"model":  DEFAULT_MODEL,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	req.Body = io.NopCloser(bytes.NewReader(jsonData))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	if res.StatusCode != 200 {
		fmt.Println("error", res.StatusCode)
	} else {
		responseToSend := ""
		defer res.Body.Close()
		bt, _ := io.ReadAll(res.Body)
		btSTR := strings.Split(string(bt), "\n")
		for _, it := range btSTR {
			fmt.Printf("\n\n %s\n", it)
			resp := map[string]interface{}{}
			json.Unmarshal([]byte(it), &resp)
			v, _ := resp["response"].(string)
			responseToSend += v
			if resp["context"] != nil {
				if ct, ok := resp["context"].([]interface{}); !ok {
					fmt.Println("unknown context")
				} else {
					for _, i := range ct {
						fmt.Println(i)
					}
				}

			}

		}

		return responseToSend
	}
	return ""
}
