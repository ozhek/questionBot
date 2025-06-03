package bot

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetQuestionsByLang(ctx context.Context, lang string) ([]Question, error) {
	questions := []Question{}

	// Fetch top-level questions (parent_id is NULL)
	rows, err := r.db.Query(ctx, "SELECT id, lang, text, answer, file_type, file_id, parent_id FROM questions WHERE lang = $1 order by id", lang)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			q  Question
			id sql.NullInt32
		)

		if err := rows.Scan(&q.ID, &q.Lang, &q.Text, &q.Answer, &q.FileType, &q.FileID, &id); err != nil {
			log.Printf("Failed to scan question: %v", err)
			continue
		}
		q.ParentID = int(id.Int32)

		// Fetch subquestions for this question
		q.SubQuestions, err = r.GetSubQuestions(ctx, q.ID)
		if err != nil {
			log.Printf("Failed to fetch subquestions for question ID %d: %v", q.ID, err)
		}

		questions = append(questions, q)
	}

	return questions, nil
}

func (r *Repository) GetSubQuestions(ctx context.Context, parentID int) ([]Question, error) {
	subQuestions := []Question{}

	rows, err := r.db.Query(ctx, "SELECT id, lang, text, answer, file_type, file_id, parent_id FROM questions WHERE parent_id = $1 order by id", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		if err := rows.Scan(&q.ID, &q.Lang, &q.Text, &q.Answer, &q.FileType, &q.FileID, &q.ParentID); err != nil {
			log.Printf("Failed to scan subquestion: %v", err)
			continue
		}
		subQuestions = append(subQuestions, q)
	}

	return subQuestions, nil
}

// SetUserLang saves or updates the user's language preference.
func (r *Repository) SetUserLang(ctx context.Context, userID int64, lang string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_languages (user_id, lang) VALUES ($1, $2)
        ON CONFLICT(user_id) DO UPDATE SET lang=excluded.lang`,
		userID, lang)
	return err
}

// GetUserLang retrieves the user's language preference, or returns "" if not set.
func (r *Repository) GetUserLang(ctx context.Context, userID int64) (string, error) {
	var lang string
	err := r.db.QueryRow(ctx, "SELECT lang FROM user_languages WHERE user_id = $1", userID).Scan(&lang)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return lang, err
}

// GetQuestionByID retrieves a question (with subquestions) by its ID.
func (r *Repository) GetQuestionByID(ctx context.Context, id int) (*Question, error) {
	var (
		q        Question
		parentID sql.NullInt32
	)
	err := r.db.QueryRow(ctx, "SELECT id, lang, text, answer, file_type, file_id, parent_id FROM questions WHERE id = $1", id).
		Scan(&q.ID, &q.Lang, &q.Text, &q.Answer, &q.FileType, &q.FileID, &parentID)
	if err != nil {
		return nil, err
	}

	q.ParentID = int(parentID.Int32)

	// Fetch subquestions for this question
	q.SubQuestions, err = r.GetSubQuestions(ctx, q.ID)
	if err != nil {
		log.Printf("Failed to fetch subquestions for question ID %d: %v", q.ID, err)
	}

	return &q, nil
}

func (r *Repository) DeleteQuestionByID(ctx context.Context, id int) error {
	qs, err := r.GetSubQuestions(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	for _, q := range qs {
		r.DeleteQuestionByID(ctx, q.ID)
	}

	_, err = r.db.Exec(ctx, "DELETE FROM questions WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

// CreateQuestion inserts a new question into the questions table.
func (r *Repository) CreateQuestion(ctx context.Context, lang, text, answer string, parentID int) (int, error) {
	row := r.db.QueryRow(
		ctx,
		"INSERT INTO questions (lang, text, answer, parent_id) VALUES ($1, $2, $3, $4) RETURNING id",
		lang, text, answer, sql.NullInt32{Int32: int32(parentID), Valid: parentID != 0},
	)
	var id int32

	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return int(id), nil
}

// UpdateQuestion updates the text and answer of a question by its ID.
func (r *Repository) UpdateQuestion(ctx context.Context, id int, text, answer string) error {
	_, err := r.db.Exec(
		ctx,
		"UPDATE questions SET text = $1, answer = $2 WHERE id = $3",
		text, answer, id,
	)
	return err
}

// UpdateQuestionFileID updates the file_id of a question by its ID.
func (r *Repository) UpdateQuestionFile(ctx context.Context, id int, fileType, fileID string) error {
	_, err := r.db.Exec(ctx, "UPDATE questions SET file_type = $1, file_id = $2 WHERE id = $3", fileType, fileID, id)
	return err
}
