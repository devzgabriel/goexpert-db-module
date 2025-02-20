package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type CurrencyHttpResponse struct {
	Bid string `json:"bid"`
}

func main() {
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer httpCancel()

	req, err := http.NewRequestWithContext(httpCtx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to execute request", "error", err)
		return
	}

	defer resp.Body.Close()

	var currencyResponse CurrencyHttpResponse
	err = json.NewDecoder(resp.Body).Decode(&currencyResponse)
	if err != nil {
		slog.Error("failed to decode response", "error", err)
		return
	}

	slog.Info("received currency response", "bid", currencyResponse.Bid)

	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open file", "error", err)
	}
	defer f.Close()

	_, err = f.WriteString("Dólar: " + currencyResponse.Bid + "\n")
	if err != nil {
		slog.Error("failed to write to file", "error", err)
	}

	slog.Info("wrote currency response to file")
}
