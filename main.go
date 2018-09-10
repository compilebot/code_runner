package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gopherpun/redis_queue"
)

// declares env variables
var (
	RedisHost     string
	ResponseQueue *redis_queue.Queue
	JobQueue      *redis_queue.Queue
)

func init() {
	RedisHost = os.Getenv("REDIS_HOST")
	ResponseQueueKey := os.Getenv("RESPONSE_QUEUE")
	JobQueueKey := os.Getenv("JOB_QUEUE")

	rq, err := redis_queue.NewQueue(RedisHost, ResponseQueueKey)
	if err != nil {
		panic(err)
	}

	jq, err := redis_queue.NewQueue(RedisHost, JobQueueKey)
	if err != nil {
		panic(err)
	}

	ResponseQueue = rq
	JobQueue = jq
}

func main() {
	pollQueue()
	select {}
}

func pollQueue() {
	rate := time.Second / 2
	throttle := time.Tick(rate)
	for newJob, err := JobQueue.PollQueue(); err == nil && newJob; {
		<-throttle
		item, err := JobQueue.Dequeue()
		if err != nil {
			// TODO
			fmt.Println("ERROR LINE 53", err)
		}
		job, _ := decodeJob(item)
		fmt.Println(job)
		response := NewBuild(job.Code, job.Language)

		ResponseQueue.Enqueue(response)
	}

}

// Job is a JSON structure representing information about the job.
type Job struct {
	ChannelID string
	Code      string
	Language  string
	RequestID string
}

func decodeJob(work string) (Job, error) {
	var job Job

	fmt.Println(work)
	err := json.Unmarshal([]byte(work), &job)

	if err != nil {
		fmt.Println(err)
	}

	return job, err
}
