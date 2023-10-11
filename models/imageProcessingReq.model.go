package models

type Visit struct {
	StoreID   string   `json:"store_id"`
	ImageURL  []string `json:"image_url"`
	VisitTime string   `json:"visit_time"`
}

type ImgProcessingRequest struct {
	Count  int     `json:"count"`
	Visits []Visit `json:"visits"`
}
