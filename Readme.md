## TO RUN THE MAIN.GO FILE FOLLOW THE BELOW STEPS

1: Installation of Go
https://go.dev/doc/install
2: cd <foldername>

```
run the command:
1: go mod init imageProcessing
2: go mod tidy
3: go run main.go

```

## ENVIRONMENT USED TO RUN THE PROJECT

OPERATING SYSTEM : MAC OS
TEXT EDITOR: VS CODE

## ASSUMPTIONS:

Assumtion1: Images will be downloaded by some api and perimeter has been recevied by them.

## DESCRIPTION

ROUTES:

1:"/api/submit/"
handle POST requests to submit a new image processing job. The function parses the request body, validates the input data, and enqueues the job for processing.

2:"/api/status"
to check the status of an image processing job. send "jobid" query parameter to retrieve the status of a specific job.

3:"/api/visits"
retrieves information about store visits within a specified date range.

## IMPLEMETAION DESCRIPTION

1: Running this program will not require any external database, it will store the data in local storage. This was most challenging to implement it could have been done more easily using mongo or sql database.

2: when the server is runned, threads will be created whose job is to listen for
the incomming job to process

3: when the job is recieved, it will be enqueued in the buffered channel for some max length and the job will be processed by threas in parellel.

4: user can check the status of the job by sending the jobid in the query parameter

5: user can check the visits by hiting the visits api (no query parameter is required as of now it is not implemeted)

## IMPROVEMENTS

1: Threads can be handled better instead of continously waiting for the job to come
threads can be created when the job is recieved and can be destroyed when the job is processed, ie something that will trigger when the job is recieved and when the job is processed.

2: Error handling can be improved

3: Query parameters are not handled for the visits api that can be done

4: The code can be refactored to make it more readable

5: MVC can be implemented to make the code more readable

6: Local storage is difficult to handle, so we can use mongo or sql database to store the data

7: Image handling should be done.
