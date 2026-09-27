package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrNotFound  = errors.New("student not found")
	ErrDuplicate = errors.New("student duplicate")
)

type StudentRepository interface {
	FindAll(
		ctx context.Context,
		search string,
		active *bool,
		minGrade *float64,
		maxGrade *float64,
		sort string,
		limit int,
		offset int,
	) ([]model.Student, int, error)

	FindAfterCursor(
		ctx context.Context,
		cursor *model.Cursor,
		limit int,
	) ([]model.Student, *model.Cursor, bool, error)

	FindByID(ctx context.Context, id string) (model.Student, error)
	Create(ctx context.Context, student model.Student) (model.Student, error)
	Update(ctx context.Context, id string, student model.Student) (model.Student, error)
	Delete(ctx context.Context, id string) error
}

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepository{db: db}
}

func normalizeID(id string) string {
	return strings.Trim(strings.TrimSpace(id), "\"'")
}

// ======================================================
// FIND ALL - OFFSET PAGINATION
// ======================================================

func (r *studentRepository) FindAll(
	ctx context.Context,
	search string,
	active *bool,
	minGrade *float64,
	maxGrade *float64,
	sortField string,
	limit int,
	offset int,
) ([]model.Student, int, error) {
	var conditions []string
	var args []any
	argNumber := 1

	if search != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("name ILIKE $%d", argNumber),
		)
		args = append(args, "%"+search+"%")
		argNumber++
	}

	if active != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("is_active = $%d", argNumber),
		)
		args = append(args, *active)
		argNumber++
	}

	if minGrade != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("grade >= $%d", argNumber),
		)
		args = append(args, *minGrade)
		argNumber++
	}

	if maxGrade != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("grade <= $%d", argNumber),
		)
		args = append(args, *maxGrade)
		argNumber++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderClause := "ORDER BY id ASC"

	switch sortField {
	case "name":
		orderClause = "ORDER BY name ASC"
	case "-name":
		orderClause = "ORDER BY name DESC"
	case "grade":
		orderClause = "ORDER BY grade ASC"
	case "-grade":
		orderClause = "ORDER BY grade DESC"
	case "nim":
		orderClause = "ORDER BY nim ASC"
	case "-nim":
		orderClause = "ORDER BY nim DESC"
	}

	countQuery := `
		SELECT COUNT(*)
		FROM students
		` + whereClause

	var total int
	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append([]any{}, args...)

	query := `
		SELECT id, nim, name, grade, is_active, owner_id
		FROM students
		` + whereClause + `
		` + orderClause + `
		LIMIT $` + fmt.Sprint(argNumber) + `
		OFFSET $` + fmt.Sprint(argNumber+1)

	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student

		if err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.OwnerID,
		); err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// ======================================================
// FIND AFTER CURSOR - CURSOR PAGINATION
// ======================================================

func (r *studentRepository) FindAfterCursor(
	ctx context.Context,
	cursor *model.Cursor,
	limit int,
) ([]model.Student, *model.Cursor, bool, error) {
	if limit < 1 {
		limit = 10
	}

	query := `
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id,
			created_at
		FROM students
	`

	args := []any{limit + 1}

	if cursor != nil {
		query += `
			WHERE (created_at, id) < ($2, $3)
		`
		args = append(args, cursor.CreatedAt, normalizeID(cursor.ID))
	}

	query += `
		ORDER BY created_at DESC, id DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, false, err
	}
	defer rows.Close()

	type studentRow struct {
		student   model.Student
		createdAt model.Cursor
	}

	results := make([]studentRow, 0, limit+1)

	for rows.Next() {
		var item studentRow

		if err := rows.Scan(
			&item.student.ID,
			&item.student.NIM,
			&item.student.Name,
			&item.student.Grade,
			&item.student.IsActive,
			&item.student.OwnerID,
			&item.createdAt.CreatedAt,
		); err != nil {
			return nil, nil, false, err
		}

		item.createdAt.ID = item.student.ID
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	hasMore := len(results) > limit

	if hasMore {
		results = results[:limit]
	}

	students := make([]model.Student, 0, len(results))
	for _, item := range results {
		students = append(students, item.student)
	}

	var nextCursor *model.Cursor

	if hasMore && len(results) > 0 {
		last := results[len(results)-1]
		nextCursor = &last.createdAt
	}

	return students, nextCursor, hasMore, nil
}

// ======================================================
// FIND BY ID
// ======================================================

func (r *studentRepository) FindByID(
	ctx context.Context,
	id string,
) (model.Student, error) {
	id = normalizeID(id)

	var student model.Student

	query := `
		SELECT id, nim, name, grade, is_active, owner_id
		FROM students
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}
	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}

// ======================================================
// CREATE
// ======================================================

func (r *studentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {
	query := `
		INSERT INTO students (
			nim, name, grade, is_active, owner_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, nim, name, grade, is_active, owner_id
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.OwnerID,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.OwnerID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, err
	}

	return result, nil
}

// ======================================================
// UPDATE
// ======================================================

func (r *studentRepository) Update(
	ctx context.Context,
	id string,
	student model.Student,
) (model.Student, error) {
	id = normalizeID(id)

	query := `
		UPDATE students
		SET nim = $1, name = $2, grade = $3, is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active, owner_id
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		id,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.OwnerID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, err
	}

	return result, nil
}

// ======================================================
// DELETE
// ======================================================

func (r *studentRepository) Delete(
	ctx context.Context,
	id string,
) error {
	id = normalizeID(id)

	result, err := r.db.Exec(
		ctx,
		`DELETE FROM students WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
