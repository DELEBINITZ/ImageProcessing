package models

import "time"

type JobStatusEnum string

const (
	StatusPending   JobStatusEnum = "OnGoing"
	StatusCompleted JobStatusEnum = "Completed"
	StatusFailed    JobStatusEnum = "Failed"
)

// Visit represents each visit data
type Visit struct {
	StoreID   string   `json:"store_id"`
	ImageURL  []string `json:"image_url"`
	VisitTime string   `json:"visit_time"`
}

// ImgProcessingRequest represents the request schema
type ImgProcessingRequest struct {
	Count  int    `json:"count"`
	Visits []Visit `json:"visits"`
}

// ImgProcessedResult represents the result schema
type ImgProcessedResult struct {
	StoreID         string
	ProcessedImages []struct {
		ImageURL  string
		Perimeter float64
	}
	VisitTime time.Time
}

// DatePerimeter represents the date and perimeter data
type DatePerimeter struct {
	Date      string  `json:"date"`
	Perimeter float64 `json:"perimeter"`
}

// VisitResponseObject represents the response schema
type VisitResponseObject struct {
	StoreID   string          `json:"store_id"`
	Area      string          `json:"area"`
	StoreName string          `json:"store_name"`
	Data      []DatePerimeter `json:"data"`
}

type Job struct {
	ID   int
	Task ImgProcessingRequest
}

type JobStatusInfo struct {
	JobID  int
	Status JobStatusEnum
	Error  []struct {
		StoreID string
		err     string
	}
}

type datePerimeter struct {
	Date      string  `json:"date"`
	Perimeter float64 `json:"perimeter"`
}

var DataBase map[string]ImgProcessedResult
