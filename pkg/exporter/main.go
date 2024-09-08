package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"math/rand"
	"net/http"
	"time"
)

// Counter类型的Metric, 用于表示单调递增的指标，例如请求数等。Counter在每次观测时会增加它所代表的值（通常是一个整数），但不会减少或者重置。
var httpRequestCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_count",  // Metric的name
		Help: "http request count"}, // Metric的说明信息
	[]string{"endpoint", "code"}) // Metric有一个Label，名称是endpoint，Metric形如 http_request_count(endpoint="")

// Gauge类型的Metric,仪表盘,用于表示可变化的度量值，例如CPU利用率、内存用量等。Gauge可以增加、减少或重置，代表着当前的状态。
var orderNum = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "order_num",
		Help: "order num"},
)

// Summary类型的Metric,类似于Histogram，也用于表示数据样本的分布情况，但同时展示更多的统计信息，如样本数量、总和、平均值、上分位数、下分位数等。
var httpRequestDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "http_request_duration",
		Help: "http request duration",
	},
	[]string{"endpoint"},
)

// 创建一个Histogram指标, 用于表示数据样本的分布情况，例如请求延迟等。
// Histogram将数据按照桶（bucket）进行划分，并对每个桶内的样本计算出一些统计信息，如样本数量、总和、平均值等。
var histogramMetric = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "example_histogram",                 // 指标名称
	Help:    "An example histogram metric.",      // 指标帮助信息
	Buckets: prometheus.LinearBuckets(0, 10, 10), // 设置桶宽度
})

// 将Metric注册到本地的Prometheus
func init() {
	prometheus.MustRegister(httpRequestCount)
	prometheus.MustRegister(orderNum)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(histogramMetric)
}

var rootCmd = &cobra.Command{
	Use: "exporter",
	Run: func(cmd *cobra.Command, args []string) {
		build()
	},
}

func main() {
	rootCmd.Execute()
}

func build() {
	// Exporter
	http.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})) // 对外暴露metrics接口，等待Prometheus来拉取
	http.HandleFunc("/hello/", hello)                                                                // 处理业务请求，并变更Metric信息
	ipport := ":8888"
	fmt.Println("服务器启动%s", ipport)
	err := http.ListenAndServe(ipport, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("process one request = %s\n", r.URL.Path)
	// Counter类型的Metric只能增
	httpRequestCount.WithLabelValues(r.URL.Path, "200").Inc()
	start := time.Now()
	n := rand.Intn(100)
	// Gauge类型的Metric可增可减
	if n >= 90 {
		orderNum.Dec()
		time.Sleep(100 * time.Millisecond)
	} else {
		orderNum.Inc()
		time.Sleep(50 * time.Millisecond)
	}
	// Summary类型Metric
	elapsed := (float64)(time.Since(start) / time.Millisecond)
	httpRequestDuration.WithLabelValues(r.URL.Path).Observe(elapsed)
	w.Write([]byte("ok"))
}
