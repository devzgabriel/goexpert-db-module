package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type CurrencyHttpResponseItem struct {
	Code string `json:"code"`
	Bid  string `json:"bid"`
}

type CurrencyHttpResponse struct {
	USDBRL CurrencyHttpResponseItem `json:"USDBRL"`
}

var sourceUrl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func main() {

	db, err := sql.Open("sqlite", "file:database.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	http.HandleFunc("/cotacao", hocHandleGetCurrency(db))

	fmt.Println("Server running on localhost:8080")
	err = http.ListenAndServe("localhost:8080", nil)

	if err != nil {
		panic(err)
	}
}

func hocHandleGetCurrency(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer httpCancel()

		req, err := http.NewRequestWithContext(httpCtx, "GET", sourceUrl, nil)
		if err != nil {
			slog.Error("failed to create request", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("failed to execute request", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		var currencyHttpResponse CurrencyHttpResponse
		err = json.NewDecoder(resp.Body).Decode(&currencyHttpResponse)
		if err != nil {
			slog.Error("failed to decode response", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("currency data", "data", currencyHttpResponse)

		dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer dbCancel()

		_, err = db.ExecContext(dbCtx, "CREATE TABLE IF NOT EXISTS currency (id INTEGER PRIMARY KEY, value REAL, created_at TEXT)")
		if err != nil {
			slog.Error("failed to create table", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = db.ExecContext(dbCtx, "INSERT INTO currency (value, created_at) VALUES (?, ?)", currencyHttpResponse.USDBRL.Bid, time.Now().Format(time.RFC3339))
		if err != nil {
			slog.Error("failed to insert currency data", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		dataR, err := json.Marshal(map[string]interface{}{
			"bid": currencyHttpResponse.USDBRL.Bid,
		})
		if err != nil {
			slog.Error("failed to marshal response", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(dataR); err != nil {
			slog.Error("failed to write json data", "error", err)
			return
		}
	}
}
