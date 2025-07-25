package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

type APIResponse struct {
	FearAndGreed struct {
		Score  float64 `json:"score"`
		Rating string  `json:"rating"`
	} `json:"fear_and_greed"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Loading .env", "err", err)
		return
	}

	getAndSendIndex()
}

func getAndSendIndex() {
	url := "https://production.dataviz.cnn.io/index/fearandgreed/graphdata/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Warn("building request", "err", err)
		return
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:140.0) Gecko/20100101 Firefox/140.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("getting response", "err", err)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("reading body", "err", err)
		return
	}

	var jsonData APIResponse
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		slog.Warn("unmarshaling data from json to go struct", "err", err)
		return
	}

	msg := fmt.Sprintf("Today's F&G Index: %.2f, %s", jsonData.FearAndGreed.Score, jsonData.FearAndGreed.Rating)
	if err := sendEmail(msg); err != nil {
		slog.Warn("sending email")
		return
	}
}

func sendEmail(message string) error {
	MailFrom := os.Getenv("MAILFROM")
	MailPass := os.Getenv("MAILPASS")
	MailTo := os.Getenv("MAILTO")
	MailHost := os.Getenv("MAILHOST")
	MailPort := os.Getenv("MAILPORT")

	addr := MailHost + ":" + MailPort
	auth := smtp.PlainAuth("", MailFrom, MailPass, MailHost)
	to := []string{MailTo}

	message = fmt.Sprintf("To: %s\nFrom: %s\nSubject: %s\nMIME-version: 1.0;\nContent-Type: text/html; charset=UTF-8;\n", MailTo, MailFrom, "F&G Index") + message

	err := smtp.SendMail(addr, auth, MailFrom, to, []byte(message))
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}

	return nil
}
