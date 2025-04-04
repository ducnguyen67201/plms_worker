package main

import (
	"code_evaluator_worker/Model"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var conn *amqp091.Connection
var channel *amqp091.Channel
var queue amqp091.Queue

func InitQueue() {
	var err error
	conn, err = amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}

	queue, err = channel.QueueDeclare(
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
}

func PublishJob(job Model.CodeJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return channel.PublishWithContext(
		ctx,
		"judge_problem",           // exchange
		queue.Name,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func ConsumeJobs(jobHandler func(Model.CodeJob)) {
	msgs, err := channel.Consume(
		"judge_problem", // queue
		"",         // consumer tag
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Fatal("Failed to consume from queue:", err)
	}

	go func() {
		for d := range msgs {
			var job Model.CodeJob
			if err := json.Unmarshal(d.Body, &job); err == nil {
				jobHandler(job)
			} else {
				log.Println("Failed to unmarshal job:", err)
			}
		}
	}()
}
