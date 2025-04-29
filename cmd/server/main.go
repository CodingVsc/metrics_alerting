package main

import (
	"net/http"
	"strconv"
	"strings"
)

// MemStorage служит для хранения метрик
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// NewMemStorage создает новый объект типа MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// MemStorageController Интерфейс для взаимодействия с хранилищем метрик
type MemStorageController interface {
	UpdateGauge(gauge string, val string) error
	UpdateCounter(counter string, val string) error
}

func (ms *MemStorage) UpdateGauge(gauge string, val string) error {
	nVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	ms.gauge[gauge] = nVal
	return nil
}

func (ms *MemStorage) UpdateCounter(counter string, val string) error {
	nVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	ms.counter[counter] += nVal
	return nil
}

func metricsController(res http.ResponseWriter, req *http.Request, storage MemStorageController) {
	if req.Method != "POST" {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Only text/plain headers are allowed!", http.StatusUnsupportedMediaType)
		return
	}
	url := req.URL.Path
	clearUrl := strings.TrimPrefix(url, "/update/")
	urlArr := strings.Split(clearUrl, "/")
	if len(urlArr) != 3 {
		http.Error(res, "Invalid URL format", http.StatusNotFound)
		return
	}
	metricType := urlArr[0]
	metricName := urlArr[1]
	metricValue := urlArr[2]

	if metricName == "" {
		http.Error(res, "Metric name is required", http.StatusNotFound)
		return
	}

	switch metricType {
	case "counter":
		err := storage.UpdateCounter(metricName, metricValue)
		if err != nil {
			http.Error(res, "Not correct value", http.StatusBadRequest)
			return
		}

	case "gauge":
		err := storage.UpdateGauge(metricName, metricValue)
		if err != nil {
			http.Error(res, "Not correct value", http.StatusBadRequest)
			return
		}
	default:
		http.Error(res, "Invalid metric type", http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
}

func main() {
	var metricsStore MemStorageController = NewMemStorage()
	http.HandleFunc("/update/", func(res http.ResponseWriter, req *http.Request) {
		metricsController(res, req, metricsStore)
	})
	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
