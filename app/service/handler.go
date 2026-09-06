package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type Handler struct {
	repo repository.StudentRepository
}

func NewHandler(repo repository.StudentRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// GET /api/v1/students
func (h *Handler) GetStudents(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	search := c.Query("search")

	var active *bool

	if value := c.Query("active"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			active = &parsed
		}
	}

	var minGrade *float64

	if value := c.Query("min_grade"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			minGrade = &parsed
		}
	}

	var maxGrade *float64

	if value := c.Query("max_grade"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			maxGrade = &parsed
		}
	}

	sortField := c.Query("sort")
	offset := (page - 1) * limit

	students, total, err := h.repo.FindAll(
		c.Context(),
		search,
		active,
		minGrade,
		maxGrade,
		sortField,
		limit,
		offset,
	)

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal mengambil data mahasiswa: "+err.Error(),
		)
	}

	totalPages := 0

	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return helper.SendSuccess(
		c,
		200,
		"Data mahasiswa berhasil diambil",
		fiber.Map{
			"items": students,
			"meta": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	)
}

// GET /api/v1/students/:id
func (h *Handler) GetStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	student, err := h.repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.SendError(
			c,
			404,
			"Data mahasiswa tidak ditemukan",
		)
	}

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal mengambil data mahasiswa: "+err.Error(),
		)
	}

	return helper.SendSuccess(
		c,
		200,
		"Data mahasiswa ditemukan",
		student,
	)
}

// POST /api/v1/students
func (h *Handler) CreateStudent(c *fiber.Ctx) error {
	if !strings.HasPrefix(
		c.Get("Content-Type"),
		"application/json",
	) {
		return helper.SendError(
			c,
			415,
			"Content-Type harus application/json",
		)
	}

	var input model.CreateStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return helper.SendError(
			c,
			400,
			"Body bukan JSON yang valid",
		)
	}

	// Business rule dipindahkan ke student_rules.go
	if err := ValidateCreateStudent(input); err != nil {
		switch {
		case errors.Is(err, ErrInvalidNIM):
			return helper.SendError(
				c,
				422,
				"Field nim wajib diisi",
			)

		case errors.Is(err, ErrInvalidName):
			return helper.SendError(
				c,
				422,
				"Field name wajib diisi",
			)

		case errors.Is(err, ErrInvalidGrade):
			return helper.SendError(
				c,
				422,
				"Grade harus berada di antara 0 sampai 100",
			)
		}
	}

	student := model.Student{
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
	}

	result, err := h.repo.Create(
		c.Context(),
		student,
	)

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.SendError(
			c,
			409,
			"NIM sudah digunakan",
		)
	}

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal menambahkan mahasiswa: "+err.Error(),
		)
	}

	c.Set(
		"Location",
		"/api/v1/students/"+result.ID,
	)

	return helper.SendSuccess(
		c,
		201,
		"Mahasiswa berhasil ditambahkan",
		result,
	)
}

// PUT /api/v1/students/:id
func (h *Handler) UpdateStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	if !strings.HasPrefix(
		c.Get("Content-Type"),
		"application/json",
	) {
		return helper.SendError(
			c,
			415,
			"Content-Type harus application/json",
		)
	}

	var input model.UpdateStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return helper.SendError(
			c,
			400,
			"Body bukan JSON yang valid",
		)
	}

	// Business rule dipindahkan ke student_rules.go
	if err := ValidateUpdateStudent(input); err != nil {
		switch {
		case errors.Is(err, ErrInvalidNIM),
			errors.Is(err, ErrInvalidName):
			return helper.SendError(
				c,
				422,
				"Field nim dan name wajib diisi untuk PUT",
			)

		case errors.Is(err, ErrInvalidGrade):
			return helper.SendError(
				c,
				422,
				"Grade harus berada di antara 0 sampai 100",
			)
		}
	}

	student := model.Student{
		ID:       id,
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
	}

	result, err := h.repo.Update(
		c.Context(),
		id,
		student,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.SendError(
			c,
			404,
			"Data mahasiswa tidak ditemukan",
		)
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.SendError(
			c,
			409,
			"NIM sudah digunakan",
		)
	}

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal memperbarui data mahasiswa: "+err.Error(),
		)
	}

	return helper.SendSuccess(
		c,
		200,
		"Data mahasiswa berhasil diperbarui",
		result,
	)
}

// PATCH /api/v1/students/:id
func (h *Handler) PatchStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	if !strings.HasPrefix(
		c.Get("Content-Type"),
		"application/json",
	) {
		return helper.SendError(
			c,
			415,
			"Content-Type harus application/json",
		)
	}

	current, err := h.repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.SendError(
			c,
			404,
			"Data mahasiswa tidak ditemukan",
		)
	}

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal mengambil data mahasiswa: "+err.Error(),
		)
	}

	var input model.PatchStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return helper.SendError(
			c,
			400,
			"Body bukan JSON yang valid",
		)
	}

	// Business rule dan penerapan perubahan dipindahkan ke student_rules.go
	if err := ApplyPatchStudent(&current, input); err != nil {
		switch {
		case errors.Is(err, ErrInvalidNIM):
			return helper.SendError(
				c,
				422,
				"NIM tidak boleh kosong",
			)

		case errors.Is(err, ErrInvalidName):
			return helper.SendError(
				c,
				422,
				"Name tidak boleh kosong",
			)

		case errors.Is(err, ErrInvalidGrade):
			return helper.SendError(
				c,
				422,
				"Grade harus berada di antara 0 sampai 100",
			)
		}
	}

	result, err := h.repo.Update(
		c.Context(),
		id,
		current,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.SendError(
			c,
			404,
			"Data mahasiswa tidak ditemukan",
		)
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.SendError(
			c,
			409,
			"NIM sudah digunakan",
		)
	}

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal memperbarui data mahasiswa: "+err.Error(),
		)
	}

	return helper.SendSuccess(
		c,
		200,
		"Sebagian data mahasiswa berhasil diperbarui",
		result,
	)
}

// DELETE /api/v1/students/:id
func (h *Handler) DeleteStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	err := h.repo.Delete(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.SendError(
			c,
			404,
			"Data mahasiswa tidak ditemukan",
		)
	}

	if err != nil {
		return helper.SendError(
			c,
			500,
			"Gagal menghapus data mahasiswa: "+err.Error(),
		)
	}

	return c.SendStatus(204)
}
