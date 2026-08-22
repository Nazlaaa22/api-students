package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students = []Student{
	{
		ID:       "1",
		NIM:      "001",
		Name:     "Nazla",
		Grade:    90,
		IsActive: true,
	},
	{
		ID:       "2",
		NIM:      "002",
		Name:     "Nafisa",
		Grade:    85,
		IsActive: true,
	},
	{
		ID:       "3",
		NIM:      "003",
		Name:     "Nana",
		Grade:    80,
		IsActive: false,
	},
}

// GET /api/v1/students
func getStudents(c *fiber.Ctx) error {
	// Pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// Search berdasarkan nama, tidak membedakan huruf besar/kecil
	search := strings.ToLower(c.Query("search"))

	// Filter
	activeFilter := c.Query("active")
	minGrade, _ := strconv.ParseFloat(c.Query("min_grade", ""), 64)
	maxGrade, _ := strconv.ParseFloat(c.Query("max_grade", ""), 64)

	filtered := make([]Student, 0)

	for _, student := range students {

		if search != "" &&
			!strings.Contains(strings.ToLower(student.Name), search) {
			continue
		}

		if activeFilter != "" {
			active, err := strconv.ParseBool(activeFilter)
			if err == nil && student.IsActive != active {
				continue
			}
		}

		if c.Query("min_grade") != "" && student.Grade < minGrade {
			continue
		}

		if c.Query("max_grade") != "" && student.Grade > maxGrade {
			continue
		}

		filtered = append(filtered, student)
	}

	// Sorting
	sortField := c.Query("sort")

	switch sortField {
	case "name":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	case "-name":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name > filtered[j].Name
		})
	case "grade":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Grade < filtered[j].Grade
		})
	case "-grade":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Grade > filtered[j].Grade
		})
	case "nim":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].NIM < filtered[j].NIM
		})
	case "-nim":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].NIM > filtered[j].NIM
		})
	}

	// Pagination
	total := len(filtered)
	totalPages := (total + limit - 1) / limit

	start := (page - 1) * limit

	if start > total {
		start = total
	}

	end := start + limit
	if end > total {
		end = total
	}

	data := filtered[start:end]

	return sendSuccess(c, 200, "Data mahasiswa berhasil diambil", fiber.Map{
		"items": data,
		"meta": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GET /api/v1/students/:id
func getStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := parseID(id); err != nil {
		return sendError(c, 400, "ID harus berupa angka")
	}

	for _, student := range students {
		if student.ID == id {
			return sendSuccess(c, 200, "Data mahasiswa ditemukan", student)
		}
	}

	return sendError(c, 404, "Data mahasiswa tidak ditemukan")
}

// POST /api/v1/students
func createStudent(c *fiber.Ctx) error {
	if !strings.HasPrefix(c.Get("Content-Type"), "application/json") {
		return sendError(c, 415, "Content-Type harus application/json")
	}

	var input CreateStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return sendError(c, 400, "Body bukan JSON yang valid")
	}

	if input.NIM == "" {
		return sendError(c, 422, "Field nim wajib diisi")
	}

	if input.Name == "" {
		return sendError(c, 422, "Field name wajib diisi")
	}

	if input.Grade < 0 || input.Grade > 100 {
		return sendError(c, 422, "Grade harus berada di antara 0 sampai 100")
	}

	// Cek NIM duplikat
	for _, student := range students {
		if student.NIM == input.NIM {
			return sendError(c, 409, "NIM sudah digunakan")
		}
	}

	// Membuat ID baru
	newID := 1

	for _, student := range students {
		id, err := strconv.Atoi(student.ID)
		if err == nil && id >= newID {
			newID = id + 1
		}
	}

	// Ubah CreateStudentRequest menjadi Student
	newStudent := Student{
		ID:       strconv.Itoa(newID),
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
	}

	students = append(students, newStudent)

	c.Set("Location", fmt.Sprintf("/api/v1/students/%s", newStudent.ID))

	return sendSuccess(c, 201, "Mahasiswa berhasil ditambahkan", newStudent)
}

// PUT /api/v1/students/:id
func updateStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := parseID(id); err != nil {
		return sendError(c, 400, "ID harus berupa angka")
	}

	if !strings.HasPrefix(c.Get("Content-Type"), "application/json") {
		return sendError(c, 415, "Content-Type harus application/json")
	}

	index := -1

	for i, student := range students {
		if student.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return sendError(c, 404, "Data mahasiswa tidak ditemukan")
	}

	var input UpdateStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return sendError(c, 400, "Body bukan JSON yang valid")
	}

	if input.NIM == "" || input.Name == "" {
		return sendError(c, 422, "Field nim dan name wajib diisi untuk PUT")
	}

	if input.Grade < 0 || input.Grade > 100 {
		return sendError(c, 422, "Grade harus berada di antara 0 sampai 100")
	}

	// Cek NIM duplikat
	for i, student := range students {
		if i != index && student.NIM == input.NIM {
			return sendError(c, 409, "NIM sudah digunakan")
		}
	}

	// Update seluruh data mahasiswa
	updatedStudent := Student{
		ID:       id,
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
	}

	students[index] = updatedStudent

	return sendSuccess(c, 200, "Data mahasiswa berhasil diperbarui", updatedStudent)
}

// PATCH /api/v1/students/:id
func patchStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := parseID(id); err != nil {
		return sendError(c, 400, "ID harus berupa angka")
	}

	if !strings.HasPrefix(c.Get("Content-Type"), "application/json") {
		return sendError(c, 415, "Content-Type harus application/json")
	}

	index := -1

	for i, student := range students {
		if student.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return sendError(c, 404, "Data mahasiswa tidak ditemukan")
	}

	var input PatchStudentRequest

	if err := c.BodyParser(&input); err != nil {
		return sendError(c, 400, "Body bukan JSON yang valid")
	}

	student := students[index]

	// Update NIM jika dikirim
	if input.NIM != nil {
		if *input.NIM == "" {
			return sendError(c, 422, "NIM tidak boleh kosong")
		}

		// Cek NIM duplikat
		for i, other := range students {
			if i != index && other.NIM == *input.NIM {
				return sendError(c, 409, "NIM sudah digunakan")
			}
		}

		student.NIM = *input.NIM
	}

	// Update Name jika dikirim
	if input.Name != nil {
		if *input.Name == "" {
			return sendError(c, 422, "Name tidak boleh kosong")
		}

		student.Name = *input.Name
	}

	// Update Grade jika dikirim
	if input.Grade != nil {
		if *input.Grade < 0 || *input.Grade > 100 {
			return sendError(c, 422, "Grade harus berada di antara 0 sampai 100")
		}

		student.Grade = *input.Grade
	}

	// Update IsActive jika dikirim
	if input.IsActive != nil {
		student.IsActive = *input.IsActive
	}

	students[index] = student

	return sendSuccess(c, 200, "Sebagian data mahasiswa berhasil diperbarui", student)
}

// DELETE /api/v1/students/:id
func deleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := parseID(id); err != nil {
		return sendError(c, 400, "ID harus berupa angka")
	}

	index := -1

	for i, student := range students {
		if student.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return sendError(c, 404, "Data mahasiswa tidak ditemukan")
	}

	students = append(students[:index], students[index+1:]...)

	return c.SendStatus(204)
}
