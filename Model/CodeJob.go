package Model

import "time"

type CodeJob struct {
	JobID      string      `json:"job_id"`
	Submission SubmitProblem `json:"submission"`
}

type SubmitProblem struct {
	SubmissionID   string    `json:"submission_id"`
	UserID         int64    `json:"user_id"`
	ProblemID      int64     `json:"problem_id"`
	SubmissionDate time.Time `json:"submission_date"`
	Result         string    `json:"result"`
	Performance    string    `json:"performance"`
	Code           string    `json:"code"`
	Language       string    `json:"language"`
}