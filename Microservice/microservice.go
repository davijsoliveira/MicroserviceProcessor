package main

import (
	"MicroserviceProcessor/Processor"
	"fmt"
	"log"
	"net/http"
)

func main() {
	Processor.Store = Processor.TrafficSignalStore{
		TrafficSignals: make(map[int]*Processor.TrafficSignalData),
	}

	http.HandleFunc("/traffic", Processor.HandleTrafficSignal)
	http.HandleFunc("/traffic/info", Processor.HandleTrafficSignalInfo)
	http.HandleFunc("/trafficflow", Processor.HandleTrafficFlow)
	http.HandleFunc("/stats", Processor.HandleStats)

	go Processor.UpdateRequestRate()

	// Inicia o servidor na porta 8082
	fmt.Println("Processor Microservice Started...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
