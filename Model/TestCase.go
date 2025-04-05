package Model

type ProblemWithTestCase struct {
	ProblemID       int64      `json:"problem_id"`
	ContestID       *int64     `json:"contest_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	DifficultyLevel string     `json:"difficulty_level"`
	RepeatedTimes   int64      `json:"repeated_times"`
	Type            string     `json:"type"`
	TestCase        []TestCase `json:"test_cases"`
}

type TestCase struct {
	TestCaseID     int64   `json:"test_case_id"`
	ProblemID      int64   `json:"problem_id"`
	Input          string  `json:"input"`
	ExpectedOutput string  `json:"expected_output"`
	CreatedAt      *string `json:"created_at"`
	UpdatedAt      *string `json:"updated_at"`
	IsActive       string  `json:"is_active"`
}

type Problem struct {
	ProblemID       int64  `json:"problem_id"`
	ContestID       *int64 `json:"contest_id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	DifficultyLevel string `json:"difficulty_level"`
	RepeatedTimes   int64  `json:"repeated_times"`
	Type            string `json:"type"`
}