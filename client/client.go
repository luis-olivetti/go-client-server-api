package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var cambioDolarReal CambioDolarReal
	err = json.Unmarshal(body, &cambioDolarReal)
	if err != nil {
		panic(err)
	}

	var bid = cambioDolarReal.USDBRL.Bid
	content := "Dólar: " + bid

	err = os.WriteFile("cotacao.txt", []byte(content), 0666)
	if err != nil {
		panic(err)
	}
}
