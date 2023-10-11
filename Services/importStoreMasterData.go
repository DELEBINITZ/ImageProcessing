package services

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	database "github.com/DELEBINITZ/imageProcessing/Database"
	"github.com/DELEBINITZ/imageProcessing/models"
)

// load store master data
func LoadStoreMasterData() {

	file, err := os.Open("StoreMasterAssignment.csv")

	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		fmt.Println("Error reading records")
	}

	for _, eachrecord := range records {
		areaCode := eachrecord[0]
		storeName := eachrecord[1]
		storeID := eachrecord[2]

		database.StoreMasterDB[storeID] = models.StoreMasterData{
			StoreName: storeName,
			AreaCode:  areaCode,
		}
	}
}
