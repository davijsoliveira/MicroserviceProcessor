package main

import (
	"MicroserviceProcessor/Processor"
	"log"
	"net/http"
)

func main() {
	Processor.Store = Processor.TrafficSignalStore{
		TrafficSignals: make(map[int]*Processor.TrafficSignalData),
	}

	http.HandleFunc("/traffic", Processor.HandleTrafficSignal)
	http.HandleFunc("/traffic/info", Processor.HandleTrafficSignalInfo)
	http.HandleFunc("/stats", Processor.HandleStats)

	go Processor.UpdateRequestRate()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
