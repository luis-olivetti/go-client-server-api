package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func fetchCambio(ctx context.Context, url string) (CambioDolarReal, error) {
	var cambio CambioDolarReal

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return cambio, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return cambio, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return cambio, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return cambio, err
	}

	err = json.Unmarshal(body, &cambio)
	return cambio, err
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0666)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cambio, err := fetchCambio(ctx, "http://localhost:8080/cotacao")
	if err != nil {
		log.Println("Fetch failed. Details: " + err.Error())

		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Operation timed out (HTTP)")
		}

		return
	}

	content := "Dólar: " + cambio.USDBRL.Bid
	err = writeToFile("cotacao.txt", content)
	if err != nil {
		log.Println("Failed to write cotacao.txt")
	}
}
