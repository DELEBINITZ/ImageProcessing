package models

type Job struct {
	ID   int
	Task ImgProcessingRequest
}

type JobStatusEnum string

type JobStatusInfo struct {
	JobID  int
	Status JobStatusEnum
	Error  []struct {
		StoreID string
		err     string
	}
}
