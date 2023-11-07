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

type CambioDolarReal struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	initServer()
}

func initServer() {
	http.HandleFunc("/cotacao", BuscaCambioDolarReal)
	http.HandleFunc("/cotacao-on-db", BuscaCambioFromDB)
	http.ListenAndServe(":8080", nil)
}

func BuscaCambioDolarReal(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ctx := r.Context()

	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var cambioDolarReal CambioDolarReal
	err = json.Unmarshal(body, &cambioDolarReal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	saveCambioToDB(&cambioDolarReal)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cambioDolarReal)
}

func saveCambioToDB(cambio *CambioDolarReal) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Grava no banco de dados usando GORM
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

	result := db.WithContext(ctx).Create(&registro) // Passa uma referência do modelo para Create
	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) {
			// Se o erro for devido ao timeout
			log.Println("Operation timed out")
		} else {
			log.Println(result.Error)
		}
		return result.Error
	}
	return nil
}

func BuscaCambioFromDB(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao-on-db" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	var cambios []Cambio
	result := db.WithContext(ctx).Find(&cambios)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)

		if errors.Is(result.Error, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusGatewayTimeout)
		}
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
