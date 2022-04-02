package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type (
	gauge float64

	counter int64

	RunMetrics struct {
		Alloc         gauge
		BuckHashSys   gauge
		Frees         gauge
		GCCPUFraction gauge
		GCSys         gauge
		HeapAlloc     gauge
		HeapIdle      gauge
		HeapInuse     gauge
		HeapObjects   gauge
		HeapReleased  gauge
		HeapSys       gauge
		LastGC        gauge
		Lookups       gauge
		MCacheInuse   gauge
		MCacheSys     gauge
		MSpanInuse    gauge
		MSpanSys      gauge
		Mallocs       gauge
		NextGC        gauge
		NumForcedGC   gauge
		NumGC         gauge
		OtherSys      gauge
		PauseTotalNs  gauge
		StackInuse    gauge
		StackSys      gauge
		Sys           gauge
		TotalAlloc    gauge
		PollCount     counter
		RandomValue   gauge
	}
)

var mapka map[string]gauge
var mapcount map[string]counter

func HandleGauge(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	logrus.Info(url)
	fields := strings.Split(url, "/")
	logrus.Info(fields)
	logrus.Info(fields[4])
	a, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		logrus.Error(err)
	}
	mapka[fields[3]] = gauge(a)
	// case condition:

	// }
	logrus.Info(mapka)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
func HandleCounter(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	logrus.Info(url)
	fields := strings.Split(url, "/")
	a, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		logrus.Error(err)
	} //Pollcount =
	mapcount[fields[3]] += counter(a)
	logrus.Info(mapcount)

	// w.Write([]byte("Fuck you counter"))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// func HandleMetricsFunc(w http.ResponseWriter, r *http.Request) {
// 	url := r.URL
// 	logrus.Info(url)
// 	mapka[
// 	// fields := strings.Split(&url, "/")

// 	w.Write([]byte("Fuck you"))
// }

func main() {
	mapka = map[string]gauge{}
	mapcount = map[string]counter{}
	// var handler3 ApiHandler
	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", HandleGauge)
	mux.HandleFunc("/update/counter/", HandleCounter)

	addr := "127.0.0.1:8080"
	server := &http.Server{Handler: mux, Addr: addr}
	// http.Handle("/api/", apiHandler)
	// http.Handle("/api/auth", apiAuthHandler)
	// http.HandleFunc("/update/", HandleMetricsFunc)

	logrus.Info(server.ListenAndServe())
}
