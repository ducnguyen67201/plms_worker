package main

import (
	"code_evaluator_worker/AppConfig"
	"code_evaluator_worker/Model"
	"code_evaluator_worker/Utils"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/alexbrainman/odbc"
	"github.com/go-redis/redis/v8"
)

var Redis *redis.Client
var ctx = context.Background()

func main() {
	// * Connect to MongoDB
	redis,err  := AppConfig.ConnectRedis()
	if err != nil { 
		log.Fatal("Failed to connect to Redis:", err)
	}
	Redis = redis
	defer redis.Close()

	// * Connect to OracleBD
	connStr := "Driver={Oracle in OraDB21Home1};Dbq=localhost:1521/xe;Uid=damg7275_final;Pwd=damg7275_final;"
	db, err := sql.Open("odbc", connStr)

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("db.Ping failed:", err)
	}

	// * Connect to RabbitMQ
	mqClient, err := AppConfig.InitRabbitMQ("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}

	defer mqClient.Conn.Close()
	defer mqClient.Channel.Close()

	q, err := mqClient.Channel.QueueDeclare(
		"judge_problem", // queue name
		true,        // durable
		false,       // auto-delete
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments	
	)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}

	msgs, err := mqClient.Channel.Consume(
		q.Name,
		"",            // consumer tag
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false, 	  // no-wait
		nil,          // arguments
	)

	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	log.Print("worker stared, waiting for messages...")

	go func() { 
		for msg := range msgs {
			// log.Printf("Received a message: %s", msg.Body)
			var job Model.CodeJob
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("Error unmarshalling message: %s", err)
				continue
			}
			log.Printf("processing ..... Job ID: %s", job.JobID)
			processJob(db, job)
		}
	}()

	select {}

}


func processJob(db *sql.DB, job Model.CodeJob) {
	// ? Get test cases for the problem
	problemWithTestCase, err := GetTestCase(db, job.Submission.ProblemID)
	if err != nil {
		log.Printf("❌ Error getting test case: %s", err)
		return
	}

	// 🔁 Loop through each test case
	for _, testCase := range problemWithTestCase.TestCase {
		log.Printf("\n\n[%s] ========================= Processing Test Case %v =========================", time.Now().Format("2006-01-02 15:04:05"), testCase.TestCaseID)
		log.Printf("🔢 TestCase Input: %s\n", testCase.Input)
		log.Printf("🎯 Expected Output: %s\n", testCase.ExpectedOutput)

		var data map[string]interface{}
		err := json.Unmarshal([]byte(testCase.Input), &data)
		if err != nil {
			panic(err)
		}

		// convert input to string
		args := data["args"]
		argsBytes, err := json.Marshal(args)
		if err != nil {
			panic(err)
		}
		argsStr := string(argsBytes)
		argsStr = argsStr[1 : len(argsStr)-1] // Remove the brackets

		// Generate the complete driver script
		driverCode := Utils.GenerateDriverScript(job.Submission.Code, problemWithTestCase.MethodName, argsStr)

		// Run the job with Utils.ProcessJob
		output, err := Utils.ProcessJob(
			job.JobID,
			job.Submission.SubmissionID,
			driverCode,
			strconv.FormatInt(testCase.TestCaseID, 10),
		)
		if err != nil {
			log.Printf("❌ Job %s failed during execution: %v\n", job.JobID, err)
			continue
		}

		output = strings.TrimSpace(output)
		output = fmt.Sprintf("{\"value\":%s}", output)
		expected := strings.TrimSpace(testCase.ExpectedOutput)

		expectedNormalized, _ := NormalizeJSON(expected)
		outputNormalized, _ := NormalizeJSON(output)


		key := job.JobID
		var outputResponse Model.SubmitProblem
		// Compare outputs
		if reflect.DeepEqual(expectedNormalized, outputNormalized) {
			log.Printf("✅ Test case %d passed\n", testCase.TestCaseID)
			outputResponse = Model.SubmitProblem{
				SubmissionID:   job.Submission.SubmissionID,
				UserID:         job.Submission.UserID,
				ProblemID:      job.Submission.UserID,
				SubmissionDate: time.Now(),
				Result:         "success",
				Performance:    "GREAT PERFORMANCE",
				Code:           job.Submission.Code,
				Language: 		job.Submission.Language,
			}
		} else {
			log.Printf("❌ Test case %d failed\nExpected: %s\nGot: %s\n", testCase.TestCaseID, expected, output)
			outputResponse = Model.SubmitProblem{
				SubmissionID:   job.Submission.SubmissionID,
				UserID:         job.Submission.UserID,
				ProblemID:      job.Submission.UserID,
				SubmissionDate: time.Now(),
				Result:         "failed",
				Performance:    "GREAT PERFORMANCE",
				Code:           job.Submission.Code,
				Language: 		job.Submission.Language,
			}
		}

		// * After finish processing the testcase, update status into redis
		UpateRedis(key, &outputResponse)
	}
}

func UpateRedis(key string, value *Model.SubmitProblem) { 
	// Check if the key already exists
	existingValue, err := Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Println("Key not found. Creating new entry.")
	} else if err != nil {
		log.Fatalf("Error checking Redis key: %v", err)
	} else {
		log.Printf("Key already exists. Value: %s — Overwriting...", existingValue)
	}

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		log.Fatalf("Failed to marshal struct: %v", err)
	}

	// Set JSON to Redis with no expiration (0 = permanent)
	err = Redis.Set(ctx, key, jsonData, 0).Err()
	if err != nil {
		log.Fatalf("Failed to set value in Redis: %v", err)
	}

	log.Printf("Stored successfully under key: %s", key)
}

func SortMapByKey(input map[string]interface{}) map[string]interface{} {
	// Extract and sort keys
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create a new map and insert in sorted key order
	sortedMap := make(map[string]interface{})
	for _, k := range keys {
		sortedMap[k] = input[k]
	}
	return sortedMap
}

func EncodePythonCode(raw string) string { 
	raw = strings.ReplaceAll(raw, "\t", "\\t")
	raw = strings.ReplaceAll(raw, "\n", "\\n")
	return raw
}

func NormalizeJSON(s string) (interface{}, error) {
	s = strings.ReplaceAll(s, "True", "true")
	s = strings.ReplaceAll(s, "False", "false")
	s = strings.ReplaceAll(s, "None", "null")

	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	return result, err
}
