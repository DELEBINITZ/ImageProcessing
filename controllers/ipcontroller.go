package controllers

import (
	"github.com/DELEBINITZ/imageProcessing/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/google/uuid"
	"sync"
)

var dataBaseLock sync.Mutex
var jobStatus map[int]models.JobStatusInfo

// servHome is the controller for the home route.
func servHome(w http.ResponseWriter, r *http.Request) {
	responseObj := map[string]interface{}{
		"message": "API is Live",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(responseObj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SubmitJob is the controller for the submit job route.
func SubmitJob(w http.ResponseWriter, r *http.Request) {
	fmt.Println("running submit job function")

	// Check if the request body is empty
	if r.Body == nil {
		fmt.Println("Body is empty")
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	// Declare a new ImgProcessingRequest struct.
	var req models.ImgProcessingRequest

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

	// check if any field is missing
	for _, visit := range req.Visits {
		if visit.StoreID == "" || len(visit.ImageURL) == 0 {
			http.Error(w, "Missing fields", http.StatusBadRequest)
			return
		}
	}

	// create unique Id for the job check for uuid
	job := models.Job{ID: int(uuid.New().ID()), Task: req}
	// add job to job queue
	jobQueue <- job

	// make job status as ongoing
	jobStatus[job.ID] = models.JobStatusInfo{
		JobID:  job.ID,
		Status: models.StatusPending,
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

// GetJobStatus is the controller for the get job status route.
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

	// acquire lock on database
	dataBaseLock.Lock()
	// release lock after function execution
	defer dataBaseLock.Unlock()

	// check if job id exists in job status map
	status, exists := jobStatus[jobIdInt]

	// if job id does not exist, return 400 error
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

// GetVisitInfo is the controller for the get visit info route.
func GetVisitInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running get visit info function")

	// create visit array of type visitResponseObject
	visits := make([]models.VisitResponseObject, 0)

	// loop over database
	for storeID, imageProcessedData := range models.DataBase {

		// get processed images from database
		processedImages := imageProcessedData.ProcessedImages

		// create data array of type datePerimeter
		var data []models.DatePerimeter

		// loop over processed images and append data to data array
		for _, processedImage := range processedImages {
			data = append(data, models.DatePerimeter{
				Date:      imageProcessedData.visitTime.Format(time.RFC3339),
				Perimeter: processedImage.Perimeter,
			})
		}

		// get store master data from store master database
		storeMaseterData := storeMasterDB[storeID]

		// create visit response object
		visitResponseObject := models.VisitResponseObject{
			StoreID:   storeID,
			Area:      storeMaseterData.areaCode,
			StoreName: storeMaseterData.storeName,
			Data:      data,
		}

		// append visit response object to visits array
		visits = append(visits, visitResponseObject)
	}

	// convert visits array to JSON
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
