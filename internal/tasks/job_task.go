package tasks

import (
	"statio/internal/services"
)

type TableTask struct {
	services *services.TableService
}

func NewTableTask(services *services.TableService) *TableTask {
	return &TableTask{services}
}
