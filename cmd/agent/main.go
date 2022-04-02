package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/fatih/structs"
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

func fill(ms runtime.MemStats, rm *RunMetrics) {
	rm.Alloc = gauge(ms.Alloc)
	rm.BuckHashSys = gauge(ms.BuckHashSys)
	rm.Frees = gauge(ms.Frees)
	rm.GCCPUFraction = gauge(ms.GCCPUFraction)
	rm.GCSys = gauge(ms.GCSys)
	rm.HeapAlloc = gauge(ms.HeapAlloc)
	rm.HeapIdle = gauge(ms.HeapIdle)
	rm.HeapInuse = gauge(ms.HeapInuse)
	rm.HeapObjects = gauge(ms.HeapObjects)
	rm.HeapReleased = gauge(ms.HeapReleased)
	rm.HeapSys = gauge(ms.HeapSys)
	rm.LastGC = gauge(ms.LastGC)
	rm.Lookups = gauge(ms.Lookups)
	rm.MCacheInuse = gauge(ms.MCacheInuse)
	rm.MCacheSys = gauge(ms.MCacheSys)
	rm.MSpanInuse = gauge(ms.MSpanInuse)
	rm.MSpanSys = gauge(ms.MSpanSys)
	rm.Mallocs = gauge(ms.Mallocs)
	rm.NextGC = gauge(ms.NextGC)
	rm.NumForcedGC = gauge(ms.NumForcedGC)
	rm.NumGC = gauge(ms.NumGC)
	rm.OtherSys = gauge(ms.OtherSys)
	rm.PauseTotalNs = gauge(ms.PauseTotalNs)
	rm.StackInuse = gauge(ms.StackInuse)
	rm.StackSys = gauge(ms.StackSys)
	rm.Sys = gauge(ms.Sys)
	rm.TotalAlloc = gauge(ms.TotalAlloc)
	rm.PollCount += 1
	rand.Seed(time.Now().UnixNano())
	rm.RandomValue = gauge(rand.Float64())
}

var (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	rm := RunMetrics{}
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	logrus.SetReportCaller(true)
	logrus.Info(m.Alloc)
	fill(m, &rm)
	logrus.Infof("%+v", rm)
	// return
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exit_chan := make(chan int)
	go func() {
		for {
			s := <-signalChanel
			switch s {
			// kill -SIGHUP XXXX [XXXX - идентификатор процесса для программы]
			case syscall.SIGINT:
				fmt.Println("Signal interrupt triggered.")
				exit_chan <- 0
				// kill -SIGTERM XXXX [XXXX - идентификатор процесса для программы]
			case syscall.SIGTERM:
				fmt.Println("Signal terminte triggered.")
				exit_chan <- 0

				// kill -SIGQUIT XXXX [XXXX - идентификатор процесса для программы]
			case syscall.SIGQUIT:
				fmt.Println("Signal quit triggered.")
				exit_chan <- 0

			default:
				fmt.Println("Unknown signal.")
				exit_chan <- 1
			}
		}
	}()

	tickerFill := time.NewTicker(pollInterval)
	go func() {
		for {
			<-tickerFill.C
			runtime.ReadMemStats(&m)
			fill(m, &rm)
		}
	}()

	tickerSendMetrics := time.NewTicker(reportInterval)
	go func() {
		for {
			<-tickerSendMetrics.C
			sendMetrics(rm)
		}
	}()

	// }
	// runtime.ReadMemStats()
	exitCode := <-exit_chan
	//stoping ticker
	logrus.Warn("Stopping tickerFill")
	tickerFill.Stop()
	logrus.Warn("Stopping tickerSendMetrics")
	tickerSendMetrics.Stop()
	logrus.Warn("Exiting with code ", exitCode)
	os.Exit(exitCode)

}

func sendMetrics(rm RunMetrics) {
	// в формате: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	// адрес сервиса (как его писать, расскажем в следующем уроке)
	rmMap := structs.Map(rm)
	val := reflect.ValueOf(rm)
	for i := 0; i < val.NumField(); i++ {
		endpoint := "http://127.0.0.1:8080/update"
		// fmt.Println(rm)
		// logrus.Info(val.Type().Field(i).Type.Name())
		// logrus.Info(val.Type().Field(i).Name)
		// logrus.Info(rmMap[val.Type().Field(i).Name])
		// logrus.Info(fmt.Sprintf("%v", rmMap[val.Type().Field(i).Name]))
		// logrus.Info(reflect.TypeOf(rmMap[val.Type().Field(i).Name]))
		endpoint = fmt.Sprintf("%s/%s/%s/%v", endpoint, val.Type().Field(i).Type.Name(), val.Type().Field(i).Name, rmMap[val.Type().Field(i).Name])
		// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
		client := &http.Client{}

		request, err := http.NewRequest(http.MethodPost, endpoint, nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		request.Header.Add("Content-Type", "text/plain")
		// request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
		response, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// печатаем код ответа
		fmt.Println("Статус-код ", response.Status)
		defer response.Body.Close()
		// читаем поток из тела ответа

		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// и печатаем его
		fmt.Println(string(body))
	}
	logrus.Info(rm.Alloc, rm.PollCount, rm.RandomValue)
}
