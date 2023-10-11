package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	channels "github.com/DELEBINITZ/imageProcessing/channels"
	database "github.com/DELEBINITZ/imageProcessing/database"
	models "github.com/DELEBINITZ/imageProcessing/models"
	"github.com/google/uuid"
)

const (
	StatusPending   models.JobStatusEnum = "OnGoing"
	StatusCompleted models.JobStatusEnum = "Completed"
	StatusFailed    models.JobStatusEnum = "Failed"
)

// SubmitJob function to submit job handels submit job route
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

	// check if any feild is missing
	for _, visit := range req.Visits {
		if visit.StoreID == "" || len(visit.ImageURL) == 0 {
			http.Error(w, "Missing feilds", http.StatusBadRequest)
			return
		}
	}

	// create unique Id for the job check for uuid
	job := models.Job{ID: int(uuid.New().ID()), Task: req}
	// add job to job queue
	channels.JobQueue <- job

	// maket job status as ongoing
	database.JobStatus[job.ID] = models.JobStatusInfo{
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
