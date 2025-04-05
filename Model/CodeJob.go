package Model

type CodeJob struct {
	JobID      string `json:"job_id"`
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
}