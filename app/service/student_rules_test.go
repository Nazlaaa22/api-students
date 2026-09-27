package service

import (
	"testing"

	"api-students/app/model"
)

func TestApplyPatchStudent(t *testing.T) {
	current := model.Student{
		ID:       "test-id",
		NIM:      "001",
		Name:     "Nazla",
		Grade:    80,
		IsActive: true,
		OwnerID:  1,
	}

	newName := "Nazla Updated"
	newGrade := 95.0
	newIsActive := false

	input := model.PatchStudentRequest{
		Name:     &newName,
		Grade:    &newGrade,
		IsActive: &newIsActive,
	}

	ApplyPatchStudent(&current, input)

	if current.Name != "Nazla Updated" {
		t.Errorf("Name tidak berubah: got %s", current.Name)
	}

	if current.Grade != 95 {
		t.Errorf("Grade tidak berubah: got %v", current.Grade)
	}

	if current.IsActive {
		t.Error("IsActive seharusnya berubah menjadi false")
	}

	if current.NIM != "001" {
		t.Errorf("NIM seharusnya tetap 001, got %s", current.NIM)
	}

	if current.OwnerID != 1 {
		t.Errorf("OwnerID seharusnya tetap 1, got %d", current.OwnerID)
	}
}

func TestApplyPatchStudentEmpty(t *testing.T) {
	current := model.Student{
		ID:       "test-id",
		NIM:      "001",
		Name:     "Nazla",
		Grade:    80,
		IsActive: true,
		OwnerID:  1,
	}

	input := model.PatchStudentRequest{}

	ApplyPatchStudent(&current, input)

	if current.Name != "Nazla" || current.Grade != 80 ||
		current.IsActive != true || current.NIM != "001" {
		t.Error("Field seharusnya tidak berubah ketika PATCH kosong")
	}
}
