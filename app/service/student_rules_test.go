package service

import (
	"errors"
	"testing"

	"api-students/app/model"
)

func TestValidateCreateStudentValid(t *testing.T) {
	input := model.CreateStudentRequest{
		NIM:      "001",
		Name:     "Nazla",
		Grade:    90,
		IsActive: true,
	}

	err := ValidateCreateStudent(input)

	if err != nil {
		t.Fatalf("data valid seharusnya tidak menghasilkan error, tetapi mendapat: %v", err)
	}
}

func TestValidateCreateStudentInvalidNIM(t *testing.T) {
	input := model.CreateStudentRequest{
		NIM:      "",
		Name:     "Nazla",
		Grade:    90,
		IsActive: true,
	}

	err := ValidateCreateStudent(input)

	if !errors.Is(err, ErrInvalidNIM) {
		t.Fatalf("seharusnya ErrInvalidNIM, tetapi mendapat: %v", err)
	}
}

func TestValidateCreateStudentInvalidGrade(t *testing.T) {
	input := model.CreateStudentRequest{
		NIM:      "001",
		Name:     "Nazla",
		Grade:    101,
		IsActive: true,
	}

	err := ValidateCreateStudent(input)

	if !errors.Is(err, ErrInvalidGrade) {
		t.Fatalf("seharusnya ErrInvalidGrade, tetapi mendapat: %v", err)
	}
}

func TestValidateUpdateStudentValid(t *testing.T) {
	input := model.UpdateStudentRequest{
		NIM:      "002",
		Name:     "Nisa",
		Grade:    85,
		IsActive: true,
	}

	err := ValidateUpdateStudent(input)

	if err != nil {
		t.Fatalf("data valid seharusnya tidak menghasilkan error, tetapi mendapat: %v", err)
	}
}

func TestApplyPatchStudent(t *testing.T) {
	current := model.Student{
		ID:       "test-id",
		NIM:      "001",
		Name:     "Nazla",
		Grade:    80,
		IsActive: true,
	}

	newName := "Nazla Updated"
	newGrade := 95.0

	input := model.PatchStudentRequest{
		Name:  &newName,
		Grade: &newGrade,
	}

	err := ApplyPatchStudent(&current, input)

	if err != nil {
		t.Fatalf("PATCH valid seharusnya tidak menghasilkan error: %v", err)
	}

	if current.Name != "Nazla Updated" {
		t.Fatalf("Name tidak berubah dengan benar")
	}

	if current.Grade != 95 {
		t.Fatalf("Grade tidak berubah dengan benar")
	}
}

func TestApplyPatchStudentInvalidGrade(t *testing.T) {
	current := model.Student{
		ID:       "test-id",
		NIM:      "001",
		Name:     "Nazla",
		Grade:    80,
		IsActive: true,
	}

	invalidGrade := 101.0

	input := model.PatchStudentRequest{
		Grade: &invalidGrade,
	}

	err := ApplyPatchStudent(&current, input)

	if !errors.Is(err, ErrInvalidGrade) {
		t.Fatalf("seharusnya ErrInvalidGrade, tetapi mendapat: %v", err)
	}
}
