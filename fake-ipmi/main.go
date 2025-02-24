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

import "os"

var nodeName string

func init() {
	var err error
	nodeName, err = os.Hostname()
	if err != nil {
		log.Fatal("Failed to get hostname:", err)
	}
}

var (
	// Node information metric
	nodeUnameInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "node_uname_info",
			Help: "Labeled system information as provided by the uname system call.",
		},
		[]string{"nodename"},
	)

	fanSpeedRPM = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_fan_speed_rpm",
			Help: "Fan speed in rotations per minute.",
		},
		[]string{"id", "name"},
	)

	fanSpeedState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_fan_speed_state",
			Help: "Reported state of a fan speed sensor (0=nominal, 1=warning, 2=critical).",
		},
		[]string{"id", "name"},
	)

	temperatureCelsius = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_temperature_celsius",
			Help: "Temperature reading in degree Celsius.",
		},
		[]string{"id", "name"},
	)

	temperatureState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_temperature_state",
			Help: "Reported state of a temperature sensor (0=nominal, 1=warning, 2=critical).",
		},
		[]string{"id", "name"},
	)

	powerWatts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_power_watts",
			Help: "Power reading in Watts.",
		},
		[]string{"id", "name"},
	)

	powerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_power_state",
			Help: "Reported state of a power sensor (0=nominal, 1=warning, 2=critical).",
		},
		[]string{"id", "name"},
	)

	currentAmperes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_current_amperes",
			Help: "Current reading in Amperes.",
		},
		[]string{"id", "name"},
	)

	currentState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_current_state",
			Help: "Reported state of a current sensor (0=nominal, 1=warning, 2=critical).",
		},
		[]string{"id", "name"},
	)

	voltageVolts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_voltage_volts",
			Help: "Voltage reading in Volts.",
		},
		[]string{"id", "name"},
	)

	voltageState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_voltage_state",
			Help: "Reported state of a voltage sensor (0=nominal, 1=warning, 2=critical).",
		},
		[]string{"id", "name"},
	)

	sensorState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_sensor_state",
			Help: "Indicates the severity of the state reported by an IPMI sensor (0=nominal, 1=warning, 2=critical).",
		},
		[]string{"id", "name", "type"},
	)

	sensorValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_sensor_value",
			Help: "Generic data read from an IPMI sensor of unknown type, relying on labels for context.",
		},
		[]string{"id", "name", "type"},
	)

	selFreeSpace = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ipmi_sel_free_space_bytes",
			Help: "Current free space remaining for new SEL entries.",
		},
	)

	selLogsCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ipmi_sel_logs_count",
			Help: "Current number of log entries in the SEL.",
		},
	)

	scrapeDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ipmi_scrape_duration_seconds",
			Help: "Returns how long the scrape took to complete in seconds.",
		},
	)

	ipmiUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ipmi_up",
			Help: "'1' if a scrape of the IPMI device was successful, '0' otherwise.",
		},
		[]string{"collector"},
	)
)

func init() {
	// Register all metrics
	prometheus.MustRegister(
		nodeUnameInfo,
		fanSpeedRPM,
		fanSpeedState,
		temperatureCelsius,
		temperatureState,
		powerWatts,
		powerState,
		currentAmperes,
		currentState,
		voltageVolts,
		voltageState,
		sensorState,
		sensorValue,
		selFreeSpace,
		selLogsCount,
		scrapeDuration,
		ipmiUp,
	)
}

func randomInRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func updateMetrics() {
	for {
		startTime := time.Now()

		// Set node information
		nodeUnameInfo.WithLabelValues(nodeName).Set(1)

		// Fan speed metrics (36 fans, 18 pairs A/B)
		for i := 1; i <= 18; i++ {
			idA := fmt.Sprintf("%d", i+3)  // IDs 4-21 for A fans
			idB := fmt.Sprintf("%d", i+21) // IDs 22-39 for B fans
			
			// A fans run faster (5880-6120 RPM)
			fanSpeedRPM.WithLabelValues(idA, fmt.Sprintf("Fan%dA", i)).Set(5880 + rand.Float64()*240)
			fanSpeedState.WithLabelValues(idA, fmt.Sprintf("Fan%dA", i)).Set(0)
			
			// B fans run slower (5040-5520 RPM)
			fanSpeedRPM.WithLabelValues(idB, fmt.Sprintf("Fan%dB", i)).Set(5040 + rand.Float64()*480)
			fanSpeedState.WithLabelValues(idB, fmt.Sprintf("Fan%dB", i)).Set(0)
		}

		// Temperature sensors
		temps := []struct{ id, name string }{
			{"1", "Temp"},
			{"2", "Temp"},
			{"3", "Inlet Temp"},
			{"171", "GPU21 Temp"},
			{"172", "GPU22 Temp"},
			{"173", "GPU23 Temp"},
			{"174", "GPU24 Temp"},
			{"180", "Exhaust Temp"},
		}
		for _, temp := range temps {
			var value float64
			switch temp.name {
			case "Inlet Temp":
				value = 21 + rand.Float64()*2
			case "Exhaust Temp":
				value = 31 + rand.Float64()*3
			case "Temp":
				value = 54 + rand.Float64()*3
			default: // GPU temps
				value = 39 + rand.Float64()*2
			}
			temperatureCelsius.WithLabelValues(temp.id, temp.name).Set(value)
			temperatureState.WithLabelValues(temp.id, temp.name).Set(0)
		}

		// Power consumption (around 1160W)
		powerWatts.WithLabelValues("91", "Pwr Consumption").Set(1160 + randomInRange(-20, 20))
		powerState.WithLabelValues("91", "Pwr Consumption").Set(0)

		// Current sensors
		currents := []struct{ id, name, value string }{
			{"81", "Current 1", "1.6"},
			{"82", "Current 2", "0.2"},
			{"251", "Current 3", "1.6"},
			{"252", "Current 4", "1.6"},
		}
		for _, current := range currents {
			baseValue := 1.6
			if current.name == "Current 2" {
				baseValue = 0.2
			}
			currentAmperes.WithLabelValues(current.id, current.name).Set(baseValue + randomInRange(-0.05, 0.05))
			currentState.WithLabelValues(current.id, current.name).Set(0)
		}

		// Voltage sensors
		voltages := []struct{ id, name string }{
			{"303", "VCORE VR"},
			{"304", "VCORE VR"},
			{"305", "MEMABCD VR"},
			{"306", "MEMEFGH VR"},
			{"307", "MEMABCD VR"},
			{"308", "MEMEFGH VR"},
			{"83", "Voltage 1"},
			{"84", "Voltage 2"},
			{"253", "Voltage 3"},
			{"254", "Voltage 4"},
		}
		for _, voltage := range voltages {
			var value float64
			switch {
			case voltage.name == "VCORE VR":
				value = randomInRange(1.18, 1.20)
			case voltage.name == "MEMABCD VR" || voltage.name == "MEMEFGH VR":
				value = randomInRange(1.21, 1.22)
			default: // Main voltages
				value = randomInRange(238, 242)
			}
			voltageVolts.WithLabelValues(voltage.id, voltage.name).Set(value)
			voltageState.WithLabelValues(voltage.id, voltage.name).Set(0)
		}

		// SEL metrics
		selFreeSpace.Set(15632)
		selLogsCount.Set(47)

		// IPMI up status
		ipmiUp.WithLabelValues("ipmi").Set(1)
		ipmiUp.WithLabelValues("sel").Set(1)

		// Scrape duration
		duration := time.Since(startTime).Seconds()
		scrapeDuration.Set(duration)

		time.Sleep(15 * time.Second)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	go updateMetrics()

	http.Handle("/metrics", promhttp.Handler())
	
	fmt.Printf("Starting IPMI metrics exporter for node %s on :9290\n", nodeName)
	fmt.Println("Access metrics at http://localhost:9290/metrics")
	log.Fatal(http.ListenAndServe(":9290", nil))
}
