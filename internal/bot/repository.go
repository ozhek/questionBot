package bot

import (
	"database/sql"
	"errors"
	"log"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetQuestionsByLang(lang string) ([]Question, error) {
	questions := []Question{}

	// Fetch top-level questions (parent_id is NULL)
	rows, err := r.db.Query("SELECT id, lang, text, answer, parent_id FROM questions WHERE lang = ?", lang)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		var id sql.NullInt32

		if err := rows.Scan(&q.ID, &q.Lang, &q.Text, &q.Answer, &id); err != nil {
			log.Printf("Failed to scan question: %v", err)
			continue
		}
		q.ParentID = int(id.Int32)

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

	rows, err := r.db.Query("SELECT id, lang, text, answer, parent_id FROM questions WHERE parent_id = ?", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.Lang, &q.Text, &q.Answer, &q.ParentID); err != nil {
			log.Printf("Failed to scan subquestion: %v", err)
			continue
		}
		subQuestions = append(subQuestions, q)
	}

	return subQuestions, nil
}

// SetUserLang saves or updates the user's language preference.
func (r *Repository) SetUserLang(userID int64, lang string) error {
	_, err := r.db.Exec(`
        INSERT INTO user_languages (user_id, lang) VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET lang=excluded.lang
    `, userID, lang)
	return err
}

// GetUserLang retrieves the user's language preference, or returns "" if not set.
func (r *Repository) GetUserLang(userID int64) (string, error) {
	var lang string
	err := r.db.QueryRow("SELECT lang FROM user_languages WHERE user_id = ?", userID).Scan(&lang)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return lang, err
}

// GetQuestionByID retrieves a question (with subquestions) by its ID.
func (r *Repository) GetQuestionByID(id int) (*Question, error) {
	var q Question
	var parentID sql.NullInt32
	err := r.db.QueryRow("SELECT id, lang, text, answer, parent_id FROM questions WHERE id = ?", id).
		Scan(&q.ID, &q.Lang, &q.Text, &q.Answer, &parentID)
	if err != nil {
		return nil, err
	}
	q.ParentID = int(parentID.Int32)

	// Fetch subquestions for this question
	q.SubQuestions, err = r.GetSubQuestions(q.ID)
	if err != nil {
		log.Printf("Failed to fetch subquestions for question ID %d: %v", q.ID, err)
	}

	return &q, nil
}

func (r *Repository) DeleteQuestionByID(id int) error {
	qs, err := r.GetSubQuestions(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	for _, q := range qs {
		r.DeleteQuestionByID(q.ID)
	}

	_, err = r.db.Exec("DELETE FROM questions WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

// CreateQuestion inserts a new question into the questions table.
func (r *Repository) CreateQuestion(lang, text, answer string, parentID int) error {
	_, err := r.db.Exec(
		"INSERT INTO questions (lang, text, answer, parent_id) VALUES (?, ?, ?, ?)",
		lang, text, answer, parentID,
	)
	return err
}

// UpdateQuestion updates the text and answer of a question by its ID.
func (r *Repository) UpdateQuestion(id int, text, answer string) error {
	_, err := r.db.Exec(
		"UPDATE questions SET text = ?, answer = ? WHERE id = ?",
		text, answer, id,
	)
	return err
}
