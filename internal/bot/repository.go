package bot

import (
	"database/sql"
	"log"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetQuestions() ([]Question, error) {
	questions := []Question{}

	// Fetch top-level questions (parent_id is NULL)
	rows, err := r.db.Query("SELECT id, text, answer FROM questions WHERE parent_id IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.Text, &q.Answer); err != nil {
			log.Printf("Failed to scan question: %v", err)
			continue
		}

		// Fetch subquestions for this question
		q.SubQuestions, err = r.GetSubQuestions(q.ID)
		if err != nil {
			log.Printf("Failed to fetch subquestions for question ID %d: %v", q.ID, err)
		}

		questions = append(questions, q)
	}

	return questions, nil
}

func (r *Repository) GetSubQuestions(parentID int) ([]Question, error) {
	subQuestions := []Question{}

	rows, err := r.db.Query("SELECT id, text, answer FROM questions WHERE parent_id = ?", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.Text, &q.Answer); err != nil {
			log.Printf("Failed to scan subquestion: %v", err)
			continue
		}
		subQuestions = append(subQuestions, q)
	}

	return subQuestions, nil
}
