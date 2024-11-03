package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

var Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var AppName = "clean-duplicates"
var Wg sync.WaitGroup
var Done = make(chan bool)
var NumberOfFiles int

func Ntfy(title string, msg string) {
	ntfyEndPoint := os.Getenv("NTFY_URL")
	ntfyTopic := os.Getenv("NTFY_TOPIC")
	ntfyUserName := os.Getenv("NTFY_USERNAME")
	ntfyPassword := os.Getenv("NTFY_PASSWORD")
	ntfyPost := fmt.Sprintf("%s/%s", ntfyEndPoint, ntfyTopic)

	Logger.Info("ntfy viper settings", slog.Any("ntfyurl", ntfyEndPoint), slog.Any("ntfytopic", ntfyTopic), slog.Any("combined", ntfyPost))
	client := http.Client{}

	req, err := http.NewRequest("POST", ntfyPost, strings.NewReader(msg))
	if err != nil {

		Logger.Error("ntfy request error", slog.Any("error", err))
	}
	req.Header.Set("Title", title)
	req.SetBasicAuth(ntfyUserName, ntfyPassword)
	resp, err := client.Do(req)
	if err != nil {

		Logger.Error("error client ntfy", slog.Any("error", err))

	}
	Logger.Info("ntfy msg sent", slog.Any("msg", msg), slog.Any("response", resp.Body), slog.Any("status", resp.Status), slog.Any("status_code", resp.StatusCode))
}
