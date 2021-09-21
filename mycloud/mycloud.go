package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CloudState struct {
	DBs []CloudDB
	VMs []VM
}

type CloudDB struct {
	Id   uint32 `json:"id"`
	Load int    `json:"load"`
	Tags tags   `json:"tags"`
}

type VM struct {
	Id   uint32 `json:"id"`
	Cost int    `json:"load"`
	Tags tags   `json:"tags"`
}

type tags map[string]string

type Billing struct {
	DBCost float32
}

var globalCloudState = CloudState{}
var globalBilling = Billing{DBCost: 1.234}
var dbTagsList = []string{}
var vmTagsList = []string{}

// DB operations
func createDB(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().ID()
	globalCloudState.DBs = append(globalCloudState.DBs, CloudDB{Id: id, Load: 10})
	prometheus.Unregister(exporter)
	fmt.Fprintf(w, "New State is %+v\n", globalCloudState)
	log.Println("New database has been created")
}

func setDBState(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var newCloudDB CloudDB
		err := json.NewDecoder(r.Body).Decode(&newCloudDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var check_id bool = false
		var index int = 0
		for ; index < len(globalCloudState.DBs); index++ {
			if globalCloudState.DBs[index].Id == newCloudDB.Id {
				check_id = true
				break
			}
		}
		if !check_id {
			http.Error(w, fmt.Sprintf("No DB with id %v\n", newCloudDB.Id), http.StatusNotFound)
			return
		}
		globalCloudState.DBs[index].Load = newCloudDB.Load
		globalCloudState.DBs[index].Tags = newCloudDB.Tags
		for tagKey := range newCloudDB.Tags {
			dbTagsList = append(dbTagsList, tagKey)
		}
		prometheus.Unregister(exporter)
		fmt.Fprintf(w, "New State is %+v\n", globalCloudState)
		log.Printf("Database %v state has been changed", newCloudDB.Id)
	} else {
		http.Error(w, "Only POST requests are acceptable\n", http.StatusMethodNotAllowed)
		return
	}
}

func getDBs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(globalCloudState.DBs)
	log.Println("DBs list have been obtained")
}

// VM operations
func createVM(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().ID()
	globalCloudState.VMs = append(globalCloudState.VMs, VM{Id: id, Cost: 5})
	prometheus.Unregister(exporter)
	fmt.Fprintf(w, "New State is %+v\n", globalCloudState)
	log.Println("New VN has been created")
}

func setVMState(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var newVM VM
		err := json.NewDecoder(r.Body).Decode(&newVM)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var check_id bool = false
		var index int = 0
		for ; index < len(globalCloudState.VMs); index++ {
			if globalCloudState.VMs[index].Id == newVM.Id {
				check_id = true
				break
			}
		}
		if !check_id {
			http.Error(w, fmt.Sprintf("No DB with id %v\n", newVM.Id), http.StatusNotFound)
			return
		}
		globalCloudState.VMs[index].Cost = newVM.Cost
		globalCloudState.VMs[index].Tags = newVM.Tags
		for tagKey := range newVM.Tags {
			vmTagsList = append(vmTagsList, tagKey)
		}
		prometheus.Unregister(exporter)
		fmt.Fprintf(w, "New State is %+v\n", globalCloudState)
		log.Printf("VM %v state has been changed", newVM.Id)
	} else {
		http.Error(w, "Only POST requests are acceptable\n", http.StatusMethodNotAllowed)
		return
	}
}

func getVMs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(globalCloudState.VMs)
	log.Println("VMs list have been obtained")
}

// Billing operations
func getCurrentBilling(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(globalBilling)
}

func setNewBilling(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := json.NewDecoder(r.Body).Decode(&globalBilling)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		fmt.Fprintf(w, "New Billing is %+v\n", globalBilling)
		log.Printf("New billing has been set: %+v", globalBilling)
	} else {
		http.Error(w, "Only POST requests are acceptable\n", http.StatusMethodNotAllowed)
		return
	}
}

type Exporter struct {
	descriptions map[uint32]*prometheus.Desc
}

const namespace = "mycloud"

var DBs []prometheus.Desc

var (
	dbCost = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "db_cost"),
		"How much my db costs", nil, nil,
	)
)

func NewExporter() *Exporter {
	return &Exporter{descriptions: make(map[uint32]*prometheus.Desc)}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for i := 0; i < len(globalCloudState.DBs); i++ {
		var tagsKeys []string
		for key := range globalCloudState.DBs[i].Tags {
			tagsKeys = append(tagsKeys, key)
		}
		sort.Strings(tagsKeys)
		e.descriptions[globalCloudState.DBs[i].Id] = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", fmt.Sprintf("db_%v_current_iops", globalCloudState.DBs[i].Id)),
			fmt.Sprintf("How loaded DB with id: %v", globalCloudState.DBs[i].Id),
			tagsKeys, nil,
		)
		log.Printf("Creating Description for DB #%v", i)
		ch <- e.descriptions[globalCloudState.DBs[i].Id]
	}
	for i := 0; i < len(globalCloudState.VMs); i++ {
		var tagsKeys []string
		for key := range globalCloudState.VMs[i].Tags {
			tagsKeys = append(tagsKeys, key)
		}
		sort.Strings(tagsKeys)
		e.descriptions[globalCloudState.VMs[i].Id] = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", fmt.Sprintf("vm_%v_current_cost", globalCloudState.VMs[i].Id)),
			fmt.Sprintf("How much costs VM with id: %v", globalCloudState.VMs[i].Id),
			tagsKeys, nil,
		)
		log.Printf("Creating Description for VM #%v", i)
		ch <- e.descriptions[globalCloudState.VMs[i].Id]
	}
	log.Println("Creating Description for DB cost")
	ch <- dbCost
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for i := 0; i < len(globalCloudState.DBs); i++ {
		var tagsKeys []string
		for key := range globalCloudState.DBs[i].Tags {
			tagsKeys = append(tagsKeys, key)
		}
		sort.Strings(tagsKeys)
		var tagValues []string
		for key := range tagsKeys {
			tagValues = append(tagValues, globalCloudState.DBs[i].Tags[tagsKeys[key]])
		}
		log.Printf("Collecting Metrics for DB #%v", i)
		ch <- prometheus.MustNewConstMetric(
			e.descriptions[globalCloudState.DBs[i].Id], prometheus.GaugeValue, float64(globalCloudState.DBs[i].Load), tagValues...,
		)
	}
	for i := 0; i < len(globalCloudState.VMs); i++ {
		var tagsKeys []string
		for key := range globalCloudState.VMs[i].Tags {
			tagsKeys = append(tagsKeys, key)
		}
		sort.Strings(tagsKeys)
		var tagValues []string
		for key := range tagsKeys {
			tagValues = append(tagValues, globalCloudState.VMs[i].Tags[tagsKeys[key]])
		}
		log.Printf("Collecting Metrics for VM #%v", i)
		ch <- prometheus.MustNewConstMetric(
			e.descriptions[globalCloudState.VMs[i].Id], prometheus.GaugeValue, float64(globalCloudState.VMs[i].Cost), tagValues...,
		)
	}
	ch <- prometheus.MustNewConstMetric(
		dbCost, prometheus.GaugeValue, float64(globalBilling.DBCost),
	)
	log.Println("Endpoint scraped")
}

var exporter *Exporter = NewExporter()

func main() {
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/createDB", createDB)
	http.HandleFunc("/setDBState", setDBState)
	http.HandleFunc("/getDBs", getDBs)

	http.HandleFunc("/createVM", createVM)
	http.HandleFunc("/setVMState", setVMState)
	http.HandleFunc("/getVMs", getVMs)

	http.HandleFunc("/getCurrentBilling", getCurrentBilling)
	http.HandleFunc("/setNewBilling", setNewBilling)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
