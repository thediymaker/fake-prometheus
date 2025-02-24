package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GPU Configuration
const (
	NUM_GPUS = 4
	TOTAL_MEMORY = 81920 // 80GB for A100
)

type metricInfo struct {
	name string
	help string
	metricType string
	labels []string
}

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Register all metrics
	registerMetrics()

	// Start updating metrics in background
	go updateMetrics()

	// Expose metrics on /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	
	fmt.Println("Starting fake GPU metrics exporter on :9400")
	fmt.Println("Access metrics at http://localhost:9400/metrics")
	log.Fatal(http.ListenAndServe(":9400", nil))
}

var metrics = map[string]*prometheus.GaugeVec{}
var counters = map[string]*prometheus.CounterVec{}

var commonLabels = []string{
	"gpu",
	"UUID",
	"device",
	"modelName",
	"Hostname",
	"DCGM_FI_DRIVER_VERSION",
}

var gpuInfo = []map[string]string{
	{
		"gpu": "0",
		"UUID": "GPU-10ac97a8-6854-4d04-4b34-354b379055b8",
		"device": "nvidia0",
		"modelName": "NVIDIA A100-SXM4-80GB",
		"Hostname": "g001",
		"DCGM_FI_DRIVER_VERSION": "560.35.03",
	},
	{
		"gpu": "1",
		"UUID": "GPU-a3194300-a020-e3ba-ac84-a37dc730aeb8",
		"device": "nvidia1",
		"modelName": "NVIDIA A100-SXM4-80GB",
		"Hostname": "g001",
		"DCGM_FI_DRIVER_VERSION": "560.35.03",
	},
	{
		"gpu": "2",
		"UUID": "GPU-3e59d793-a4c9-8da2-093c-716183e7049a",
		"device": "nvidia2",
		"modelName": "NVIDIA A100-SXM4-80GB",
		"Hostname": "g001",
		"DCGM_FI_DRIVER_VERSION": "560.35.03",
	},
	{
		"gpu": "3",
		"UUID": "GPU-890b2d19-ed6f-15b1-3f1b-bdd34a7fa7c6",
		"device": "nvidia3",
		"modelName": "NVIDIA A100-SXM4-80GB",
		"Hostname": "g001",
		"DCGM_FI_DRIVER_VERSION": "560.35.03",
	},
}

var metricDefinitions = []metricInfo{
	{"DCGM_FI_DEV_SM_CLOCK", "SM clock frequency (in MHz).", "gauge", commonLabels},
	{"DCGM_FI_DEV_MEM_CLOCK", "Memory clock frequency (in MHz).", "gauge", commonLabels},
	{"DCGM_FI_DEV_MEMORY_TEMP", "Memory temperature (in C).", "gauge", commonLabels},
	{"DCGM_FI_DEV_GPU_TEMP", "GPU temperature (in C).", "gauge", commonLabels},
	{"DCGM_FI_DEV_POWER_USAGE", "Power draw (in W).", "gauge", commonLabels},
	{"DCGM_FI_DEV_TOTAL_ENERGY_CONSUMPTION", "Total energy consumption since boot (in mJ).", "counter", commonLabels},
	{"DCGM_FI_DEV_GPU_UTIL", "GPU utilization (in %).", "gauge", commonLabels},
	{"DCGM_FI_DEV_MEM_COPY_UTIL", "Memory utilization (in %).", "gauge", commonLabels},
	{"DCGM_FI_DEV_ENC_UTIL", "Encoder utilization (in %).", "gauge", commonLabels},
	{"DCGM_FI_DEV_DEC_UTIL", "Decoder utilization (in %).", "gauge", commonLabels},
	{"DCGM_FI_DEV_FB_FREE", "Frame buffer memory free (in MB).", "gauge", commonLabels},
	{"DCGM_FI_DEV_FB_USED", "Frame buffer memory used (in MB).", "gauge", commonLabels},
	{"DCGM_FI_DEV_PCIE_REPLAY_COUNTER", "Total number of PCIe retries.", "counter", commonLabels},
}

func registerMetrics() {
	for _, metric := range metricDefinitions {
		if metric.metricType == "gauge" {
			metrics[metric.name] = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: metric.name,
					Help: metric.help,
				},
				metric.labels,
			)
			prometheus.MustRegister(metrics[metric.name])
		} else if metric.metricType == "counter" {
			counters[metric.name] = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: metric.name,
					Help: metric.help,
				},
				metric.labels,
			)
			prometheus.MustRegister(counters[metric.name])
		}
	}
}

func randomInRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func updateMetrics() {
	// Initialize energy consumption with random starting values
	energyConsumption := make(map[string]float64)
	for _, gpu := range gpuInfo {
		energyConsumption[gpu["gpu"]] = randomInRange(1200000000000, 1800000000000)
	}

	for {
		for _, gpu := range gpuInfo {
			gpuNum := gpu["gpu"]
			
			// Randomly determine if GPU is in use (80% chance if previously in use)
			isActive := rand.Float64() < 0.8

			// GPU Utilization (0-100%)
			gpuUtil := 0.0
			if isActive {
				gpuUtil = randomInRange(60, 100) // High utilization when active
			} else {
				gpuUtil = randomInRange(0, 15) // Low utilization when idle
			}
			metrics["DCGM_FI_DEV_GPU_UTIL"].With(gpu).Set(gpuUtil)

			// Memory Clock (1593 MHz for A100, slight variations)
			metrics["DCGM_FI_DEV_MEM_CLOCK"].With(gpu).Set(randomInRange(1590, 1595))

			// SM Clock (210-1410 MHz)
			if gpuUtil > 50 {
				metrics["DCGM_FI_DEV_SM_CLOCK"].With(gpu).Set(randomInRange(1380, 1410))
			} else {
				metrics["DCGM_FI_DEV_SM_CLOCK"].With(gpu).Set(randomInRange(210, 300))
			}

			// Temperatures correlate with utilization but have some randomness
			baseTemp := 30.0 + (gpuUtil * 0.5)
			metrics["DCGM_FI_DEV_GPU_TEMP"].With(gpu).Set(baseTemp + randomInRange(-2, 2))
			metrics["DCGM_FI_DEV_MEMORY_TEMP"].With(gpu).Set(baseTemp + randomInRange(-5, 10))

			// Power usage correlates with utilization (60W idle, up to 440W max)
			basePower := 60 + (380 * gpuUtil / 100)
			metrics["DCGM_FI_DEV_POWER_USAGE"].With(gpu).Set(basePower + randomInRange(-10, 10))

			// Memory utilization and usage
			memUtil := 0.0
			if isActive {
				memUtil = randomInRange(10, 90)
			} else {
				memUtil = randomInRange(0, 5)
			}
			metrics["DCGM_FI_DEV_MEM_COPY_UTIL"].With(gpu).Set(memUtil)

			usedMem := (TOTAL_MEMORY * memUtil / 100)
			metrics["DCGM_FI_DEV_FB_USED"].With(gpu).Set(usedMem)
			metrics["DCGM_FI_DEV_FB_FREE"].With(gpu).Set(TOTAL_MEMORY - usedMem)

			// Encoder/Decoder (usually 0 for compute cards, but occasionally show small values)
			metrics["DCGM_FI_DEV_ENC_UTIL"].With(gpu).Set(randomInRange(0, 1))
			metrics["DCGM_FI_DEV_DEC_UTIL"].With(gpu).Set(randomInRange(0, 1))

			// Energy consumption increases over time
			energyConsumption[gpuNum] += basePower * 15 * 1000 // 15 seconds in millijoules
			counters["DCGM_FI_DEV_TOTAL_ENERGY_CONSUMPTION"].With(gpu).Add(basePower * 15 * 1000)

			// PCIe retries (very rare, only increment occasionally)
			if rand.Float64() < 0.01 { // 1% chance
				counters["DCGM_FI_DEV_PCIE_REPLAY_COUNTER"].With(gpu).Inc()
			}
		}
		time.Sleep(15 * time.Second)
	}
}
