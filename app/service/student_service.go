package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
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

// ======================================================
// HELPER AUTHORIZATION
// ======================================================

func getCurrentUser(c *fiber.Ctx) (model.AuthUser, error) {
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return model.AuthUser{}, errors.New("user_id tidak ditemukan")
	}

	username, ok := c.Locals("username").(string)
	if !ok {
		return model.AuthUser{}, errors.New("username tidak ditemukan")
	}

	role, ok := c.Locals("role").(string)
	if !ok {
		return model.AuthUser{}, errors.New("role tidak ditemukan")
	}

	return model.AuthUser{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, nil
}

func getPermissionSet(c *fiber.Ctx) helper.PermissionSet {
	role, ok := c.Locals("role").(string)

	if !ok || role == "" {
		return helper.PermissionSet{}
	}

	return helper.PermissionsForRole(role)
}

// ======================================================
// CONTENT NEGOTIATION
// ======================================================

// negotiateStudentsFormat menentukan format response berdasarkan
// header Accept. Jika Accept kosong atau */*, format default JSON.
func negotiateStudentsFormat(accept string) string {
	accept = strings.TrimSpace(strings.ToLower(accept))

	if accept == "" || accept == "*/*" {
		return "json"
	}

	for _, item := range strings.Split(accept, ",") {
		parts := strings.Split(item, ";")
		mediaType := strings.TrimSpace(parts[0])

		// Abaikan media type yang memiliki q=0.
		q := 1.0
		for _, param := range parts[1:] {
			param = strings.TrimSpace(param)

			if strings.HasPrefix(param, "q=") {
				value, err := strconv.ParseFloat(
					strings.TrimSpace(strings.TrimPrefix(param, "q=")),
					64,
				)
				if err != nil {
					q = 0
				} else {
					q = value
				}
			}
		}

		if q <= 0 {
			continue
		}

		switch mediaType {
		case "application/json", "application/*":
			return "json"
		case "text/csv", "text/*":
			return "csv"
		case "*/*":
			return "json"
		}
	}

	return ""
}

// ======================================================
// GET /api/v1/students
// JSON & CSV CONTENT NEGOTIATION
// ======================================================

func (h *Handler) GetStudents(c *fiber.Ctx) error {
	// --------------------------------------------------
	// NEGOSIASI FORMAT RESPONSE
	// --------------------------------------------------

	format := negotiateStudentsFormat(c.Get("Accept"))

	if format == "" {
		return helper.SendError(
			c,
			fiber.StatusNotAcceptable,
			"Format response tidak didukung. Gunakan Accept: application/json atau text/csv",
		)
	}

	// --------------------------------------------------
	// PAGINATION
	// --------------------------------------------------

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	var cursor *model.Cursor

	if encoded := c.Query("cursor"); encoded != "" {
		decoded, err := helper.DecodeCursor(encoded)
		if err != nil {
			return helper.BadRequest("Cursor tidak valid")
		}

		cursor = &decoded
	}

	// --------------------------------------------------
	// AMBIL DATA
	// --------------------------------------------------

	students, nextCursor, hasMore, err := h.repo.FindAfterCursor(
		c.Context(),
		cursor,
		limit,
	)

	if err != nil {
		return helper.Internal(err)
	}

	// --------------------------------------------------
	// RESPONSE CSV
	// --------------------------------------------------

	if format == "csv" {
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)

		// Header CSV
		if err := writer.Write([]string{
			"id",
			"nim",
			"name",
			"grade",
			"is_active",
			"owner_id",
		}); err != nil {
			return helper.Internal(err)
		}

		// Data mahasiswa
		for _, student := range students {
			if err := writer.Write([]string{
				student.ID,
				student.NIM,
				student.Name,
				fmt.Sprint(student.Grade),
				strconv.FormatBool(student.IsActive),
				strconv.Itoa(student.OwnerID),
			}); err != nil {
				return helper.Internal(err)
			}
		}

		writer.Flush()

		if err := writer.Error(); err != nil {
			return helper.Internal(err)
		}

		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set(
			"Content-Disposition",
			"attachment; filename=students.csv",
		)

		return c.Send(buffer.Bytes())
	}

	// --------------------------------------------------
	// RESPONSE JSON
	// --------------------------------------------------

	meta := fiber.Map{
		"limit":    limit,
		"has_more": hasMore,
	}

	if nextCursor != nil {
		meta["next_cursor"] = helper.EncodeCursor(
			nextCursor.CreatedAt,
			nextCursor.ID,
		)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"Data mahasiswa berhasil diambil",
		fiber.Map{
			"items": students,
			"meta":  meta,
		},
	)
}

// ======================================================
// GET /api/v1/students/:id
// ======================================================

func (h *Handler) GetStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	student, err := h.repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if err != nil {
		return helper.Internal(err)
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		return helper.Unauthorized("User belum terautentikasi")
	}

	permissions := getPermissionSet(c)

	if !CanAccessStudent(
		currentUser,
		student.OwnerID,
		&permissions,
		"student:read:any",
	) {
		return helper.Forbidden(
			"Tidak memiliki akses ke data mahasiswa ini",
		)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"Data mahasiswa ditemukan",
		student,
	)
}

// ======================================================
// POST /api/v1/students
// ======================================================

func (h *Handler) CreateStudent(c *fiber.Ctx) error {
	if !strings.HasPrefix(
		c.Get("Content-Type"),
		"application/json",
	) {
		return helper.UnsupportedMediaType(
			"Content-Type harus application/json",
		)
	}

	var input model.CreateStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return helper.BadRequest("Body bukan JSON yang valid")
	}

	if validationErrors := helper.ValidateStruct(input); validationErrors != nil {
		return helper.Validation(validationErrors)
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		return helper.Unauthorized("User belum terautentikasi")
	}

	student := model.Student{
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
		OwnerID:  currentUser.UserID,
	}

	result, err := h.repo.Create(
		c.Context(),
		student,
	)

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Conflict("NIM sudah digunakan")
	}

	if err != nil {
		return helper.Internal(err)
	}

	c.Set(
		"Location",
		"/api/v1/students/"+result.ID,
	)

	return helper.SendSuccess(
		c,
		fiber.StatusCreated,
		"Mahasiswa berhasil ditambahkan",
		result,
	)
}

// ======================================================
// PUT /api/v1/students/:id
// ======================================================

func (h *Handler) UpdateStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	if !strings.HasPrefix(
		c.Get("Content-Type"),
		"application/json",
	) {
		return helper.UnsupportedMediaType(
			"Content-Type harus application/json",
		)
	}

	existing, err := h.repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if err != nil {
		return helper.Internal(err)
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		return helper.Unauthorized("User belum terautentikasi")
	}

	permissions := getPermissionSet(c)

	if !CanAccessStudent(
		currentUser,
		existing.OwnerID,
		&permissions,
		"student:update:any",
	) {
		return helper.Forbidden(
			"Tidak memiliki akses untuk memperbarui data mahasiswa ini",
		)
	}

	var input model.UpdateStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return helper.BadRequest("Body bukan JSON yang valid")
	}

	if validationErrors := helper.ValidateStruct(input); validationErrors != nil {
		return helper.Validation(validationErrors)
	}

	student := model.Student{
		ID:       id,
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
		OwnerID:  existing.OwnerID,
	}

	result, err := h.repo.Update(
		c.Context(),
		id,
		student,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Conflict("NIM sudah digunakan")
	}

	if err != nil {
		return helper.Internal(err)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"Data mahasiswa berhasil diperbarui",
		result,
	)
}

// ======================================================
// PATCH /api/v1/students/:id
// ======================================================

func (h *Handler) PatchStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	if !strings.HasPrefix(
		c.Get("Content-Type"),
		"application/json",
	) {
		return helper.UnsupportedMediaType(
			"Content-Type harus application/json",
		)
	}

	current, err := h.repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if err != nil {
		log.Printf("PATCH student FindByID (id=%q): %v", id, err)
		return helper.Internal(err)
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		return helper.Unauthorized("User belum terautentikasi")
	}

	permissions := getPermissionSet(c)

	if !CanAccessStudent(
		currentUser,
		current.OwnerID,
		&permissions,
		"student:update:any",
	) {
		return helper.Forbidden(
			"Tidak memiliki akses untuk memperbarui data mahasiswa ini",
		)
	}

	var input model.PatchStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return helper.BadRequest("Body bukan JSON yang valid")
	}

	if validationErrors := helper.ValidateStruct(input); validationErrors != nil {
		return helper.Validation(validationErrors)
	}

	ApplyPatchStudent(&current, input)

	result, err := h.repo.Update(
		c.Context(),
		id,
		current,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Conflict("NIM sudah digunakan")
	}

	if err != nil {
		log.Printf("PATCH student Update (id=%q): %v", id, err)
		return helper.Internal(err)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"Sebagian data mahasiswa berhasil diperbarui",
		result,
	)
}

// ======================================================
// DELETE /api/v1/students/:id
// ======================================================

func (h *Handler) DeleteStudent(c *fiber.Ctx) error {
	id := strings.Trim(c.Params("id"), "\"'")

	existing, err := h.repo.FindByID(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if err != nil {
		return helper.Internal(err)
	}

	currentUser, err := getCurrentUser(c)
	if err != nil {
		return helper.Unauthorized("User belum terautentikasi")
	}

	permissions := getPermissionSet(c)

	if !CanAccessStudent(
		currentUser,
		existing.OwnerID,
		&permissions,
		"student:delete:any",
	) {
		return helper.Forbidden(
			"Tidak memiliki akses untuk menghapus data mahasiswa ini",
		)
	}

	err = h.repo.Delete(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound("Data mahasiswa tidak ditemukan")
	}

	if err != nil {
		return helper.Internal(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
