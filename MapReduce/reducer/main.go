package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func main() {
	http.HandleFunc("/reduce", reduceHandler)
	log.Println("Reducer running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func reduceHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Bucket  string   `json:"bucket"`
		Results []string `json:"results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create S3 session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))
	svc := s3.New(sess)

	// Aggregate all mapper results
	finalCount := make(map[string]int)

	for _, resultKey := range request.Results {
		result, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(request.Bucket),
			Key:    aws.String(resultKey),
		})
		if err != nil {
			continue
		}

		content, _ := io.ReadAll(result.Body)
		result.Body.Close()

		var mapperResult map[string]int
		json.Unmarshal(content, &mapperResult)

		// Merge counts
		for word, count := range mapperResult {
			finalCount[word] += count
		}
	}

	// Sort by count (top 20 words)
	var sorted []WordCount
	for word, count := range finalCount {
		sorted = append(sorted, WordCount{Word: word, Count: count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	// Prepare final result
	finalResult := map[string]interface{}{
		"total_unique_words": len(finalCount),
		"top_20_words":       sorted[:min(20, len(sorted))],
		"all_words":          finalCount,
		"timestamp":          time.Now().Unix(),
	}

	// Save to S3
	finalKey := fmt.Sprintf("final/result_%d.json", time.Now().Unix())
	jsonData, _ := json.MarshalIndent(finalResult, "", "  ")

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(finalKey),
		Body:   bytes.NewReader(jsonData),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return summary
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result_location": finalKey,
		"unique_words":    len(finalCount),
		"top_5_words":     sorted[:min(5, len(sorted))],
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
