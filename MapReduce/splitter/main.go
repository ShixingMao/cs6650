package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	http.HandleFunc("/split", splitHandler)
	log.Println("Splitter running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func splitHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.URL.Query().Get("bucket")
	fileName := r.URL.Query().Get("file")

	// Create S3 session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))
	svc := s3.New(sess)

	// Download file from S3
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	// Read content
	content, _ := io.ReadAll(result.Body)
	text := string(content)

	// Split into 3 chunks
	words := strings.Fields(text)
	chunkSize := len(words) / 3
	var chunks []string

	for i := 0; i < 3; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == 2 {
			end = len(words)
		}
		chunk := strings.Join(words[start:end], " ")

		// Upload chunk to S3
		chunkKey := fmt.Sprintf("chunks/chunk_%d.txt", i)
		_, err = svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(chunkKey),
			Body:   bytes.NewReader([]byte(chunk)),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		chunks = append(chunks, chunkKey)
	}

	// Return chunk locations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chunks": chunks,
	})
}
