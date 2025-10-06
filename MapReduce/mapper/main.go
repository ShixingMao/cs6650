package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	http.HandleFunc("/map", mapHandler)
	log.Println("Mapper running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mapHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.URL.Query().Get("bucket")
	chunkKey := r.URL.Query().Get("chunk")

	// Create S3 session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))
	svc := s3.New(sess)

	// Download chunk from S3
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(chunkKey),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	// Count words
	content, _ := io.ReadAll(result.Body)
	words := strings.Fields(strings.ToLower(string(content)))
	wordCount := make(map[string]int)

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:'\"")
		if word != "" {
			wordCount[word]++
		}
	}

	// Save result to S3
	resultKey := strings.Replace(chunkKey, "chunks/", "mapped/", 1)
	resultKey = strings.Replace(resultKey, ".txt", ".json", 1)

	jsonData, _ := json.Marshal(wordCount)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(resultKey),
		Body:   bytes.NewReader(jsonData),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return result location
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"result": resultKey,
		"status": "success",
	})
}
