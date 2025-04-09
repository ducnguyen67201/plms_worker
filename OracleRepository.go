package main

import (
	"code_evaluator_worker/Model"
	"database/sql"
	"log"
)

func GetTestCase(db *sql.DB, problem_id int64) (*Model.ProblemWithTestCase, error) {
	rows, err := db.Query(`SELECT * FROM ProblemWithTestCases WHERE problem_id = :1`, problem_id)

	if err != nil {
		log.Fatal("error executing query:", err)
		return nil, err
	}
	var problem *Model.ProblemWithTestCase
	var firstRow bool = true 
	defer rows.Close()
	for rows.Next() { 
		var ( 
			problemID       int64
			contestID       *int64
			title           string
			description     string
			difficultyLevel string
			repeatedTimes   int64
			problemType     string
			methodName 		string
			skeletonCode 	string

			testCaseID 	int64
			input 		string
			expectedOutput string
			createdAt 	*string
			updatedAt 	*string
			isActive 	string
		)

		err := rows.Scan(
			&problemID,
			&contestID,
			&title,
			&description,
			&difficultyLevel,
			&repeatedTimes,
			&problemType,
			&methodName,
			&skeletonCode,

			&testCaseID,
			&input,
			&expectedOutput,
			&createdAt,
			&updatedAt,
			&isActive,
		)
		if err != nil {
			log.Fatal("error scanning row:", err)
		}

		if firstRow { 
			problem = &Model.ProblemWithTestCase{
				ProblemID:       problemID,
				ContestID:       contestID,
				Title:           title,
				Description:     description,
				DifficultyLevel: difficultyLevel,
				RepeatedTimes:   repeatedTimes,
				Type:            problemType,
				MethodName:      methodName,
				SkeletonCode:    skeletonCode,
				TestCase:       []Model.TestCase{},
			}
			firstRow = false
		}

		problem.TestCase = append(problem.TestCase, Model.TestCase{
			TestCaseID:     testCaseID,
			Input:          input,
			ExpectedOutput: expectedOutput,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			IsActive:       isActive,
		})
	}

	return problem, nil
}