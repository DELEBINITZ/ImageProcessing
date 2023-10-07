package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MaxQueueSize = 5
	Threads      = 2
)

// TODO: time format
type Visit struct {
	StoreID   string   `json:"store_id"`
	ImageURL  []string `json:"image_url"`
	VisitTime string   `json:"visit_time"`
}

// put validation method in structs
type ImgProcessingRequest struct {
	Count  int     `json:"count"`
	Visits []Visit `json:"visits"`
}

type ImgProcessedResult struct {
	StoreID         string
	ProcessedImages []struct {
		ImageURL  string
		Perimeter float64
	}
	visitTime string
}

type JobStatusEnum string

const (
	StatusPending   JobStatusEnum = "OnGoing"
	StatusCompleted JobStatusEnum = "Completed"
	StatusFailed    JobStatusEnum = "Failed"
)

var JobStatus map[int]JobStatusInfo

type JobStatusInfo struct {
	JobID  int
	Status JobStatusEnum
	Error  []struct {
		StoreID string
		err     string
	}
}

var DataBase map[string]ImgProcessedResult
var dataBaseLock sync.Mutex
var jobStatus map[int]JobStatusInfo

type Job struct {
	ID   int
	Task ImgProcessingRequest
}

var jobQueue chan Job

func main() {
	fmt.Println("Running Image Processing API")
	port := ":8080"

	jobQueue = make(chan Job, MaxQueueSize)
	DataBase = make(map[string]ImgProcessedResult)
	jobStatus = make(map[int]JobStatusInfo)

	for i := 1; i <= Threads; i++ {
		go createThreads(i)
	}

	http.HandleFunc("/", servHome)
	http.HandleFunc("/api/submit/", SubmitJob)
	http.HandleFunc("/api/status", GetJobStatus)
	http.HandleFunc("/api/visits", GetVisitInfo)

	http.ListenAndServe(port, nil)
}

func servHome(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("API is Live")

	responseObj := map[string]interface{}{
		"message": "API is Live",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(responseObj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SubmitJob(w http.ResponseWriter, r *http.Request) {
	fmt.Println("running submit job function")

	// Check if the request body is empty
	if r.Body == nil {
		fmt.Println("Body is empty")
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	var req ImgProcessingRequest

	// Decode the JSON request body into 'req'
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if 'req.Count' is the same as the number of 'req.Visits'
	if req.Count != len(req.Visits) {
		http.Error(w, "Count is not equal to the number of visits", http.StatusBadRequest)
		return
	}

	// check if 0 count and 0 visits 400 error
	if req.Count == 0 && len(req.Visits) == 0 {
		http.Error(w, "Count and visits are empty", http.StatusBadRequest)
		return
	}

	// check if any feild is missing
	for _, visit := range req.Visits {
		if visit.StoreID == "" || len(visit.ImageURL) == 0 {
			http.Error(w, "Missing feilds", http.StatusBadRequest)
			return
		}
	}

	// create unique Id for the job check for uuid
	job := Job{ID: int(uuid.New().ID()), Task: req}
	jobQueue <- job

	// maket job status as ongoing
	jobStatus[job.ID] = JobStatusInfo{
		JobID:  job.ID,
		Status: StatusPending,
	}

	responseObject := map[string]interface{}{
		"job_id": job.ID,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(responseObject); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("jobid")
	if jobId == "" {
		http.Error(w, "Missing job id parameter", http.StatusBadRequest)
		return
	}

	jobIdInt, err := strconv.Atoi(jobId)
	if err != nil {
		http.Error(w, "{}", http.StatusBadRequest)
		return
	}

	dataBaseLock.Lock()
	defer dataBaseLock.Unlock()

	status, exists := jobStatus[jobIdInt]

	if !exists {
		http.Error(w, "Job not found", http.StatusBadRequest)
		return
	}

	responseObject := map[string]interface{}{
		"status": status,
		"job_id": jobIdInt,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(responseObject); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetVisitInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running get visit info function")

	storeId := r.URL.Query().Get("storeid")
	if storeId == "" {
		http.Error(w, "Missing store id parameter", http.StatusBadRequest)
		return
	}

	startDate := r.URL.Query().Get("startdate")
	if startDate == "" {
		http.Error(w, "Missing start date parameter", http.StatusBadRequest)
		return
	}

	endDate := r.URL.Query().Get("enddate")
	if endDate == "" {
		http.Error(w, "Missing end date parameter", http.StatusBadRequest)
		return
	}

}

func createThreads(id int) {
	for {
		select {
		case job, ok := <-jobQueue:
			if !ok {
				fmt.Println("Channel is closed")
				return
			}

			ProcessImages(job.Task, job.ID)
		}
	}
}

func ProcessImages(req ImgProcessingRequest, jobId int) {
	fmt.Println("Running Process images function")

	for _, visit := range req.Visits {

		// check if store id is empty
		if visit.StoreID == "" {
			fmt.Println("Store ID is empty")
			jobStatus[jobId] = JobStatusInfo{
				JobID:  jobId,
				Status: StatusFailed,
				Error: []struct {
					StoreID string
					err     string
				}{
					{
						StoreID: visit.StoreID,
						err:     "Store ID is empty",
					},
				},
			}
		}

		var storeId string = visit.StoreID

		var processedData []struct {
			ImageURL  string
			Perimeter float64
		}

		for _, imgUrl := range visit.ImageURL {
			time.Sleep(10 * time.Second)
			perimenter := calculatePerimeter(imgUrl)

			processedData = append(processedData, struct {
				ImageURL  string
				Perimeter float64
			}{ImageURL: imgUrl, Perimeter: float64(perimenter)})

		}

		precessedResult := ImgProcessedResult{
			StoreID:         storeId,
			ProcessedImages: processedData,
			visitTime:       visit.VisitTime,
		}

		// save processed result in database
		persistData(storeId, precessedResult)
	}
	// make job status as completed
	fmt.Println("Job completed")
	jobStatus[jobId] = JobStatusInfo{
		JobID:  jobId,
		Status: StatusCompleted,
	}
}

func calculatePerimeter(imgUrl string) int {
	fmt.Println("Running calculate perimeter function")
	return rand.Intn(100)
}

func persistData(storeID string, result ImgProcessedResult) {
	dataBaseLock.Lock()
	defer dataBaseLock.Unlock()
	DataBase[storeID] = result
	return
}
