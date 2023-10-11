package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	database "github.com/DELEBINITZ/imageProcessing/database"
	models "github.com/DELEBINITZ/imageProcessing/models"
	services "github.com/DELEBINITZ/imageProcessing/services"
	"github.com/google/uuid"
)

const (
	MaxQueueSize = 5
	Threads      = 2
)

// stores each visit data
type Visit struct {
	StoreID   string   `json:"store_id"`
	ImageURL  []string `json:"image_url"`
	VisitTime string   `json:"visit_time"`
}

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
	visitTime time.Time
}

type DataPerimeter struct {
	Date      string  `json:"date"`
	Perimeter float64 `json:"perimeter"`
}
type VisitResponseObject struct {
	StoreID   string          `json:"store_id"`
	Area      string          `json:"area"`
	StoreName string          `json:"store_name"`
	Data      []DataPerimeter `json:"data"`
}

type JobStatusEnum string

const (
	StatusPending   JobStatusEnum = "OnGoing"
	StatusCompleted JobStatusEnum = "Completed"
	StatusFailed    JobStatusEnum = "Failed"
)

type JobStatusInfo struct {
	JobID  int
	Status JobStatusEnum
	Error  []struct {
		StoreID string
		err     string
	}
}

// key is store id and value is processed result
var DataBase map[string]ImgProcessedResult
var dataBaseLock sync.Mutex

// key is job id and value is job status
var JobStatus map[int]JobStatusInfo

type Job struct {
	ID   int
	Task ImgProcessingRequest
}

var JobQueue chan Job

func main() {
	fmt.Println("Running Image Processing API")

	// create job queue for processing images
	JobQueue = make(chan Job, MaxQueueSize)
	// create database for storing processed images
	DataBase = make(map[string]ImgProcessedResult)
	// create job status map
	JobStatus = make(map[int]JobStatusInfo)
	// create store master database
	database.StoreMasterDB = make(map[string]models.StoreMasterData)

	port := ":8080"

	// load store master data from csv using go routine
	go services.LoadStoreMasterData()

	// create threads for processing images
	for i := 1; i <= Threads; i++ {
		go createThreads(i)
	}

	// handle routes and requests
	http.HandleFunc("/", servHome)
	http.HandleFunc("/api/submit/", SubmitJob)
	http.HandleFunc("/api/status", GetJobStatus)
	http.HandleFunc("/api/visits", GetVisitInfo)

	// start server
	http.ListenAndServe(port, nil)
}

// home route handler
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

// submit job route handler
func SubmitJob(w http.ResponseWriter, r *http.Request) {
	fmt.Println("running submit job function")

	// Check if the request body is empty
	if r.Body == nil {
		fmt.Println("Body is empty")
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	// Declare a new ImgProcessingRequest struct.
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
	// add job to job queue
	JobQueue <- job

	// maket job status as ongoing
	JobStatus[job.ID] = JobStatusInfo{
		JobID:  job.ID,
		Status: StatusPending,
	}

	// create response object
	responseObject := map[string]interface{}{
		"job_id": job.ID,
	}

	// send response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(responseObject); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// get job status route handler
func GetJobStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running get job status function")

	// get the job id from query params
	jobId := r.URL.Query().Get("jobid")

	// check if job id is empty
	if jobId == "" {
		http.Error(w, "Missing job id parameter", http.StatusBadRequest)
		return
	}

	// convert jobId from string to int and check for error
	jobIdInt, err := strconv.Atoi(jobId)
	if err != nil {
		http.Error(w, "{}", http.StatusBadRequest)
		return
	}

	// aquiare lock on database
	dataBaseLock.Lock()
	// release lock after function execution
	defer dataBaseLock.Unlock()

	// check if job id exists in job status map
	status, exists := JobStatus[jobIdInt]

	// if job id does not exists return 400 error
	if !exists {
		http.Error(w, "Job not found", http.StatusBadRequest)
		return
	}

	// create response object
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

// get visit info route handler
func GetVisitInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running get visit info function")

	// create visie array of type VisitResponseObject
	visits := make([]VisitResponseObject, 0)

	// loop over database
	for storeId, imageProcessedData := range DataBase {

		// get processed images from database
		processedImages := imageProcessedData.ProcessedImages

		// create data array of type DataPerimeter
		var data []DataPerimeter

		// loop over processed images and append data to data array
		for _, processedImage := range processedImages {
			data = append(data, DataPerimeter{
				Date:      imageProcessedData.visitTime.Format(time.RFC3339),
				Perimeter: processedImage.Perimeter,
			})
		}

		// get store master data from store master database
		StoreMasterData := database.StoreMasterDB[storeId]

		// create visit response object
		VisitResponseObject := VisitResponseObject{
			StoreID:   storeId,
			Area:      StoreMasterData.AreaCode,
			StoreName: StoreMasterData.StoreName,
			Data:      data,
		}

		// append visit response object to visits array
		visits = append(visits, VisitResponseObject)
	}

	// convert visits array to json
	jsonData, err := json.Marshal(visits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Set the content type header to indicate JSON response
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response
	w.Write(jsonData)

}

func createThreads(id int) {
	for {
		select {
		case job, ok := <-JobQueue:
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

	// loop over visits array
	for _, visit := range req.Visits {

		// check if store id is empty
		if visit.StoreID == "" {
			fmt.Println("Store ID is empty")
			JobStatus[jobId] = JobStatusInfo{
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

		// loop over image url array
		for _, imgUrl := range visit.ImageURL {
			// mimic image processing
			time.Sleep(2 * time.Second)
			// mimics calculating perimeter
			perimenter := calculatePerimeter(imgUrl)

			// append processed data to processedData array
			processedData = append(processedData, struct {
				ImageURL  string
				Perimeter float64
			}{ImageURL: imgUrl, Perimeter: float64(perimenter)})

		}

		// convery string to time
		dateInTimeFormat := stringToDate(visit.VisitTime)

		// create processed result object
		precessedResult := ImgProcessedResult{
			StoreID:         storeId,
			ProcessedImages: processedData,
			visitTime:       dateInTimeFormat,
		}

		// save processed result in database
		persistData(storeId, precessedResult)
	}
	// make job status as completed
	fmt.Println("Job completed")
	JobStatus[jobId] = JobStatusInfo{
		JobID:  jobId,
		Status: StatusCompleted,
	}
}

// mimics calculating perimeter
func calculatePerimeter(imgUrl string) int {
	fmt.Println("Running calculate perimeter function")
	return rand.Intn(100)
}

// save processed result in database
func persistData(storeID string, result ImgProcessedResult) {
	// aquiare lock on database
	dataBaseLock.Lock()
	// release lock after function execution
	defer dataBaseLock.Unlock()
	// save processed result in database
	DataBase[storeID] = result
	return
}

// helper function to convert string to time
func stringToDate(date string) time.Time {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		fmt.Println(err)
	}
	return t
}
