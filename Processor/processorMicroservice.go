package Processor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type TrafficSignalData struct {
	TrafficSignal   TrafficSignal
	FlowData        []int
	AverageFlowRate int
}

type TrafficSignal struct {
	ID              int
	Congestion      int
	RedLightTime    int
	YellowLightTime int
	GreenLightTime  int
}

type TrafficSignalStore struct {
	mu             sync.Mutex
	TrafficSignals map[int]*TrafficSignalData
	ActiveRequests int64
}

var Store TrafficSignalStore
var (
	totalRequests int64
	reqPerSecond  int64
)

func HandleTrafficSignal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	atomic.AddInt64(&totalRequests, 1)
	atomic.AddInt64(&reqPerSecond, 1)

	var trafficSignal TrafficSignal
	err := json.NewDecoder(r.Body).Decode(&trafficSignal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid request payload")
		return
	}

	Store.mu.Lock()
	defer Store.mu.Unlock()

	trafficSignalData, ok := Store.TrafficSignals[trafficSignal.ID]
	if !ok {
		trafficSignalData = &TrafficSignalData{
			TrafficSignal: trafficSignal,
			FlowData:      make([]int, 0),
		}
		Store.TrafficSignals[trafficSignal.ID] = trafficSignalData
	}

	trafficSignalData.FlowData = append(trafficSignalData.FlowData, trafficSignal.Congestion)

	if len(trafficSignalData.FlowData) > 10 {
		trafficSignalData.FlowData = trafficSignalData.FlowData[len(trafficSignalData.FlowData)-10:]
	}

	total := 0
	for _, flow := range trafficSignalData.FlowData {
		total += flow
	}

	trafficSignalData.AverageFlowRate = total / len(trafficSignalData.FlowData)

	// Exibir informações da requisição POST
	fmt.Printf("Traffic Signal ID: %d, Congestion: %d, Red Light Time: %d, Yellow Light Time: %d, Green Light Time: %d\n",
		trafficSignal.ID, trafficSignal.Congestion, trafficSignal.RedLightTime, trafficSignal.YellowLightTime, trafficSignal.GreenLightTime)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Traffic signal data stored successfully")
}

func HandleTrafficSignalInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	trafficSignalID := r.URL.Query().Get("id")
	if trafficSignalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing traffic signal ID parameter")
		return
	}

	Store.mu.Lock()
	defer Store.mu.Unlock()

	trafficSignalIDInt, err := strconv.Atoi(trafficSignalID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid traffic signal ID")
		return
	}

	trafficSignalData, ok := Store.TrafficSignals[trafficSignalIDInt]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Traffic signal data not found")
		return
	}

	response := struct {
		AverageFlowRate int `json:"averageFlowRate"`
	}{
		AverageFlowRate: trafficSignalData.AverageFlowRate,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error marshaling response: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func HandleStats(w http.ResponseWriter, r *http.Request) {
	totalReq := atomic.LoadInt64(&totalRequests)
	currentReqPerSec := atomic.LoadInt64(&reqPerSecond)

	response := struct {
		TotalRequests     int64 `json:"totalRequests"`
		RequestsPerSecond int64 `json:"requestsPerSecond"`
	}{
		TotalRequests:     totalReq,
		RequestsPerSecond: currentReqPerSec,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error marshaling response: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func UpdateRequestRate() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var previousTotalRequests int64

	for range ticker.C {
		Store.mu.Lock()

		totalRequests := atomic.LoadInt64(&totalRequests)
		requestsPerSecond := totalRequests - previousTotalRequests

		previousTotalRequests = totalRequests

		Store.mu.Unlock()

		fmt.Printf("Requests per Second: %d\n", requestsPerSecond)
	}
	//ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()
	//
	//for range ticker.C {
	//	currentReqPerSec := atomic.LoadInt64(&reqPerSecond)
	//	atomic.StoreInt64(&reqPerSecond, 0)
	//
	//	fmt.Printf("Requisições por segundo: %d\n", currentReqPerSec)
	//}
}
