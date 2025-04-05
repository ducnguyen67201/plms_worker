package main

import (
	"code_evaluator_worker/AppConfig"
	"code_evaluator_worker/Model"
	"code_evaluator_worker/Utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/alexbrainman/odbc"
)


func main() {
	// * Connect to MongoDB
	redis,err  := AppConfig.ConnectRedis()
	if err != nil { 
		log.Fatal("Failed to connect to Redis:", err)
	}
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
			log.Printf("Received a message: %s", msg.Body)
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

	// Extract method name and parameter names
	// methodName := job.ProblemMeta.MethodName
	// paramNames := job.ProblemMeta.ParameterNames
	methodName := "add"
	paramNames, err := Utils.ExtractPythonMethod(job.Submission.Code)
	if err != nil {
		log.Printf("❌ Error extracting method name and parameter names: %s", err)
		return
	}

	paramNamesList := paramNames.ParamName
	if len(paramNamesList) == 0 {
		log.Printf("❌ No parameter names found")
	}

	// 🔁 Loop through each test case
	for _, testCase := range problemWithTestCase.TestCase {
		fmt.Println("\n\n========================= Processing Test Case", testCase.TestCaseID, "=========================")
		fmt.Println("🔢 TestCase Input:", testCase.Input)
		fmt.Println("🎯 Expected Output:", testCase.ExpectedOutput)

		// * pre-process input values and parsed them into parameters slices of string 
		var values []string
		if strings.Contains(testCase.Input, ",") {
			values = strings.Split(testCase.Input, ",")
		} else {
			values = strings.Fields(testCase.Input)
		}

		var parameter = make(map[string]interface{})
		for val := range values { 
			parameter[paramNamesList[val]] = values[val]
		}

		// Generate the complete driver script
		driverCode := Utils.GenerateDriverScript(job.Submission.Code, methodName, parameter)

		// Run the job with Utils.ProcessJob
		output, err := Utils.ProcessJob(
			job.JobID,
			job.Submission.SubmissionID,
			driverCode,
			methodName,
			paramNamesList,
		)
		if err != nil {
			log.Printf("❌ Job %s failed during execution: %v\n", job.JobID, err)
			continue
		}

		output = strings.TrimSpace(output)
		expected := strings.TrimSpace(testCase.ExpectedOutput)

		// Optional normalization
		normalize := func(s string) string {
			return strings.Join(strings.Fields(s), " ")
		}

		// Compare outputs
		if normalize(output) == normalize(expected) {
			log.Printf("✅ Test case %d passed\n", testCase.TestCaseID)
		} else {
			log.Printf("❌ Test case %d failed\nExpected: %s\nGot: %s\n", testCase.TestCaseID, expected, output)
		}
	}
}
