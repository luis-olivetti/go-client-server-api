package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	initServer()
}

func initServer() {
	http.HandleFunc("/cotacao", BuscaCambioDolarReal)
	http.HandleFunc("/cotacao-on-db", BuscaCambioFromDB)
	http.ListenAndServe(":8080", nil)
}

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

func saveCambioToDB(cambio *CambioDolarReal) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	registro := Cambio{
		Code:       cambio.USDBRL.Code,
		Codein:     cambio.USDBRL.Codein,
		Name:       cambio.USDBRL.Name,
		High:       cambio.USDBRL.High,
		Low:        cambio.USDBRL.Low,
		VarBid:     cambio.USDBRL.VarBid,
		PctChange:  cambio.USDBRL.PctChange,
		Bid:        cambio.USDBRL.Bid,
		Ask:        cambio.USDBRL.Ask,
		Timestamp:  cambio.USDBRL.Timestamp,
		CreateDate: cambio.USDBRL.CreateDate,
	}

	result := db.WithContext(ctx).Create(&registro)
	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) {
			log.Println("Operation timed out")
		} else {
			log.Println(result.Error)
		}
		return result.Error
	}
	return nil
}

func BuscaCambioDolarReal(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	cambio, err := fetchCambio(ctx, "https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		println(err)

		if errors.Is(err, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	saveCambioToDB(&cambio)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cambio)
}

func BuscaCambioFromDB(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao-on-db" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	var cambios []Cambio
	result := db.WithContext(ctx).Order("created_at desc").Find(&cambios)
	if result.Error != nil {
		println(result.Error)

		if errors.Is(result.Error, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(cambios) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(cambios)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
