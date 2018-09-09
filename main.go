package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
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
		job := decodeJob(item)
		fmt.Println(job.code, job.code)
		response := NewBuild(job.code, job.language)

		ResponseQueue.Enqueue(response)
	}

}

type Job struct {
	session   *discordgo.Session
	channelID string
	code      string
	language  string
}

func decodeJob(work string) Job {
	var jobGob bytes.Buffer
	var job Job
	dec := gob.NewDecoder(&jobGob)
	err := dec.Decode(&work)
	if err != nil && err.Error() != "EOF" {
		fmt.Println("ERR LINE 77", err)
	}

	fmt.Println(work)

	return job
}
