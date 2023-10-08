package main
import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"github.com/google/uuid"
	"github.com/DELEBINITZ/imageProcessing/controllers"
)

const (
	MaxQueueSize = 5
	Threads      = 2
)

// ...

var jobQueue chan controllers.Job

// load store master data
func loadStoreMasterData() {
	// ...
}

func main() {
	fmt.Println("Running Image Processing API")

	// create job queue for processing images
	jobQueue = make(chan controllers.Job, MaxQueueSize)
	// create database for storing processed images
	controllers.DataBase = make(map[string]controllers.ImgProcessedResult)
	// create job status map
	controllers.JobStatus = make(map[int]controllers.JobStatusInfo)
	// create store master database
	controllers.StoreMasterDB = make(map[string]controllers.StoreMaseterData)

	port := ":8080"

	// load store master data from CSV using a goroutine
	go loadStoreMasterData()

	// create threads for processing images
	for i := 1; i <= Threads; i++ {
		go createThreads(i)
	}

	// handle routes and requests
	http.HandleFunc("/", controllers.ServHome)
	http.HandleFunc("/api/submit/", controllers.SubmitJob)
	http.HandleFunc("/api/status", controllers.GetJobStatus)
	http.HandleFunc("/api/visits", controllers.GetVisitInfo)

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

	// loop over visits array
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
	jobStatus[jobId] = JobStatusInfo{
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
