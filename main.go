package main

import (
	"code_evaluator_worker/AppConfig"
	"code_evaluator_worker/Model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
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
			log.Printf("Job ID: %d, Language: %s, Source Code: %s", job.JobID, job.Language, job.SourceCode)
			processJob(job)
		}
	}()

	select {}

}


func processJob(job Model.CodeJob) {
	filename := fmt.Sprintf("/tmp/%s.py", job.JobID)
	if err := os.WriteFile(filename, []byte(job.SourceCode), 0644); err != nil {
		log.Printf("Failed to write file for job %s: %v\n", job.JobID, err)
		return
	}
	defer exec.Command("rm", "-f", filename).Run()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", filename)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Job %s timed out\n", job.JobID)
		// Save timeout error to DB/Redis
		return
	}

	if err != nil {
		log.Printf("Job %s failed: %v\n", job.JobID, err)
	}

	log.Printf("✅ Job %s output:\n%s\n", job.JobID, string(out))

	// Save result to Redis/DB (not shown here)
}
