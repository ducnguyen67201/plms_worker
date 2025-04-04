package Model

type CodeJob struct {
	JobID      int64  `json:"job_id"`
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
}