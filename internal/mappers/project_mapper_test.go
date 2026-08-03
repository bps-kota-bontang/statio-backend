package mappers

import (
	"statio/internal/models"
	"testing"
)

func TestToProjectResponseIncludesTableMetadata(t *testing.T) {
	project := &models.Project{
		ID:   "project-id",
		Name: "Project",
		Tables: []models.ProjectTable{{
			TableID:     "table-id",
			Page:        "12-13",
			TableNumber: "2.1",
			Table:       &models.Table{ID: "table-id", Name: "Table"},
		}},
	}

	response := ToProjectResponse(project)
	if len(response.Tables) != 1 {
		t.Fatalf("expected one project table, got %d", len(response.Tables))
	}
	if response.Tables[0].TableID != "table-id" || response.Tables[0].Page != "12-13" || response.Tables[0].TableNumber != "2.1" {
		t.Fatalf("project table metadata was not mapped: %#v", response.Tables[0])
	}
}
