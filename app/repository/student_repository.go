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

	FindByID(ctx context.Context, id string) (model.Student, error)

	Create(ctx context.Context, student model.Student) (model.Student, error)

	Update(
		ctx context.Context,
		id string,
		student model.Student,
	) (model.Student, error)

	Delete(ctx context.Context, id string) error
}

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepository{
		db: db,
	}
}


// NORMALIZE UUID
func normalizeID(id string) string {
	// Menghapus spasi dan tanda kutip jika ID
	// masuk dalam bentuk:
	// "6c357605-4d6e-4504-a291-44b8f7faaf1"

	return strings.Trim(
		strings.TrimSpace(id),
		"\"'",
	)
}


// FIND ALL
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

	// Search berdasarkan nama
	if search != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("name ILIKE $%d", argNumber),
		)

		args = append(args, "%"+search+"%")
		argNumber++
	}

	// Filter status aktif
	if active != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("is_active = $%d", argNumber),
		)

		args = append(args, *active)
		argNumber++
	}

	// Filter nilai minimum
	if minGrade != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("grade >= $%d", argNumber),
		)

		args = append(args, *minGrade)
		argNumber++
	}

	// Filter nilai maksimum
	if maxGrade != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("grade <= $%d", argNumber),
		)

		args = append(args, *maxGrade)
		argNumber++
	}

	// WHERE
	whereClause := ""

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// SORTING
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

	// COUNT
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

	// QUERY DATA
	queryArgs := append([]any{}, args...)

	query := `
		SELECT id, nim, name, grade, is_active
		FROM students
		` + whereClause + `
		` + orderClause + `
		LIMIT $` + fmt.Sprint(argNumber) + `
		OFFSET $` + fmt.Sprint(argNumber+1)

	queryArgs = append(
		queryArgs,
		limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		query,
		queryArgs...,
	)

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
		); err != nil {
			return nil, 0, err
		}

		students = append(
			students,
			student,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// ======================================================
// FIND BY ID
// ======================================================

func (r *studentRepository) FindByID(
	ctx context.Context,
	id string,
) (model.Student, error) {

	// Bersihkan ID terlebih dahulu
	id = normalizeID(id)

	var student model.Student

	query := `
		SELECT id, nim, name, grade, is_active
		FROM students
		WHERE id = $1
	`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
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
			nim,
			name,
			grade,
			is_active
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nim, name, grade, is_active
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
	)

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

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

	// Bersihkan ID terlebih dahulu
	id = normalizeID(id)

	query := `
		UPDATE students
		SET
			nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active
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
	)

	// Data tidak ditemukan
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	// Error lainnya
	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {

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

	// Bersihkan ID terlebih dahulu
	id = normalizeID(id)

	query := `
		DELETE FROM students
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	if err != nil {
		return err
	}

	// Tidak ada data yang dihapus
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
