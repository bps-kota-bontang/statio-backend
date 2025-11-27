package tasks

import (
	"context"
	"encoding/json"
	"log"
	"statio/internal/dto"
	"statio/internal/services"

	"github.com/hibiken/asynq"
)

type FactTask struct {
	services *services.FactService
}

func NewFactTask(services *services.FactService) *FactTask {
	return &FactTask{services}
}

func (t *FactTask) UnanalyzeFacts(ctx context.Context, task *asynq.Task) error {
	log.Println("Unanalyze facts task started")

	var payload dto.UnanalyzeFactPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("Unanalyzing facts for table ID: %s", payload.TableID)

	// Call the service method to unanalyze facts
	if err := t.services.UnanalyzeFacts(payload.TableID); err != nil {
		log.Printf("Error unanalyzing facts for table ID %s: %v", payload.TableID, err)
		return nil
	}

	log.Printf("Successfully unanalyzed facts for table ID: %s", payload.TableID)

	return nil
}

func (t *FactTask) AnalyzeFacts(ctx context.Context, task *asynq.Task) error {
	log.Println("Analyze facts started")

	var payload dto.AnalyzeFactPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("Analyzing facts for table ID: %s", payload.TableID)

	// Call the service method to analyze facts
	if err := t.services.AnalyzeFacts(payload.TableID); err != nil {
		log.Printf("Error analyzing facts for table ID %s: %v", payload.TableID, err)
		return nil
	}

	log.Printf("Successfully analyzed facts for table ID: %s", payload.TableID)

	return nil
}

func (t *FactTask) CommitFacts(ctx context.Context, task *asynq.Task) error {
	log.Println("Commit facts started")

	var payload dto.CommitFactPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("Committing facts for table ID: %s", payload.TableID)

	// Call the service method to commit facts
	if err := t.services.CommitFacts(payload.TableID); err != nil {
		log.Printf("Error committing facts for table ID %s: %v", payload.TableID, err)
		return nil
	}

	log.Printf("Successfully committed facts for table ID: %s", payload.TableID)

	return nil
}
