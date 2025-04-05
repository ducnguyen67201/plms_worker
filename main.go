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
	"os/exec"
	"time"

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


func processJob(db *sql.DB,job Model.CodeJob) {
	filename := fmt.Sprintf("./tmp/job_%s.py", job.JobID)

	Utils.SaveFile(filename, job.Submission.Code)
	
	defer exec.Command("rm", "-f", filename).Run()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	// ? Get test case for problem from problem_id
	problemWithTestCase, err := GetTestCase(db, job.Submission.ProblemID)
	if err != nil {
		log.Printf("Error getting test case: %s", err)
		return
	}
	fmt.Println("problemWithTestCase: ", problemWithTestCase)

	// execute the Python script 
	cmd := exec.CommandContext(ctx, "python", filename)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("job %s timed out\n", job.JobID)
		// Save timeout error to DB/Redis
		return
	}

	if err != nil {
		log.Printf("Job %s failed: %v\n", job.JobID, err)
	}

	log.Printf("✅ Job %s output:\n%s\n", job.JobID, string(out))


	Utils.RemoveFile(filename)
	// Save result to Redis/DB (not shown here)
}

