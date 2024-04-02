package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/** 
 * The error message can be any text found on the HTML pages of the website. 
 * Modify the constant accordingly to match the specific error message to identify inactive pages. 
**/
const (
	inactiveErrorMessage = "product was not found"
)

func main() {
	start := time.Now()
	inFilename := "../feed.txt"
	inFile, err := os.Open(inFilename)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	resultsFilename := "../success-list.csv"
	resultsFileInactivePageName := "../error-inactive-list.csv"
	resultsFileErrorOthersName := "../error-others-list.csv"

	resultsFile, err := os.Create(resultsFilename)
	if err != nil {
		panic(err)
	}
	defer resultsFile.Close()

	resultsFileInactive, err := os.Create(resultsFileInactivePageName)
	if err != nil {
		panic(err)
	}
	defer resultsFileInactive.Close()

	resultsFileErrorOthers, err := os.Create(resultsFileErrorOthersName)
	if err != nil {
		panic(err)
	}
	defer resultsFileErrorOthers.Close()

	taskChan := make(chan Task, 40000)
	resultChan := make(chan Result, 40000)
	workers := 8
	for w := 0; w < workers; w++ {
		go checkWorker(w+1, taskChan, resultChan)
	}

	count := 0
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if err = scanner.Err(); err != nil {
			panic(err)
		}
		fmt.Printf("main: feeding record: %+v\n", line)
		taskChan <- Task{count, line}
		count++
	}
	close(taskChan)

	resultWriter := csv.NewWriter(resultsFile)
	resultsInactiveWriter := csv.NewWriter(resultsFileInactive)
	resultsErrorOthersWriter := csv.NewWriter(resultsFileErrorOthers)

	for i := 0; i < count; i++ {
		res := <-resultChan
		if res.Valid && res.Status == 200 {
			err = resultWriter.Write([]string{res.URL, strconv.Itoa(res.Status), strconv.FormatBool(res.Valid), res.ErrorMsg})
		} else if !res.Valid && strings.Contains(res.ErrorMsg, inactiveErrorMessage) {
			err = resultsInactiveWriter.Write([]string{res.URL, strconv.Itoa(res.Status), strconv.FormatBool(res.Valid), res.ErrorMsg})
		} else {
			err = resultsErrorOthersWriter.Write([]string{res.URL, strconv.Itoa(res.Status), strconv.FormatBool(res.Valid), res.ErrorMsg})
		}

		if err != nil {
			fmt.Printf("error writing result %+v: %+v", res, err)
		}

		fmt.Printf("main: result: %+v\n", res)
	}

	close(resultChan)
	resultWriter.Flush()
	resultsInactiveWriter.Flush()
	resultsErrorOthersWriter.Flush()

	if err = resultWriter.Error(); err != nil {
		panic(err)
	}
	if err = resultsInactiveWriter.Error(); err != nil {
		panic(err)
	}
	if err = resultsErrorOthersWriter.Error(); err != nil {
		panic(err)
	}

	elapsed := time.Since(start)
	fmt.Println("Async link checking for", elapsed.Seconds(), "seconds")
}

type Task struct {
	Index int
	URL   string
}

type Result struct {
	Index    int
	URL      string
	Status   int
	Valid    bool
	ErrorMsg string
}

func checkWorker(id int, tasks <-chan Task, r chan<- Result) {
	fmt.Printf("worker %02d: spinning up\n", id)
	client := &http.Client{}
	for t := range tasks {
		fmt.Printf("worker %02d: received task %+v\n", id, t)
		result := checker(client, t)
		fmt.Printf("worker %02d: completed task with result: %+v\n", id, result)
		r <- result
	}
	fmt.Printf("worker %02d: no more tasks, exiting\n", id)
}

func checker(client *http.Client, t Task) Result {
	result := Result{
		Index:    t.Index,
		URL:      t.URL,
		Status:   -1,
		Valid:    false,
		ErrorMsg: "none",
	}
	req, err := http.NewRequest("GET", t.URL, nil)
	if err != nil {
		result.ErrorMsg = err.Error()
		return result
	}
	// Set the User-Agent header to "facebookexternalhit/1.1" to emulate Facebook's web crawler.
	req.Header.Add("User-Agent", "facebookexternalhit/1.1")

	res, err := client.Do(req)
	if err != nil {
		result.ErrorMsg = err.Error()
		return result
	}
	defer res.Body.Close()

	result.Status = res.StatusCode
	valid, errmsg := validateResponse(res, []string{inactiveErrorMessage})
	result.Valid = valid
	result.ErrorMsg = errmsg

	return result
}

func validateResponse(resp *http.Response, errorMessages []string) (bool, string) {
	for _, errMsg := range errorMessages {
		if bytes.Contains([]byte(resp.Status), []byte(errMsg)) {
			return false, errMsg
		}
	}
	return true, "none"
}
