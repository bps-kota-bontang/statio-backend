package di

import (
	"statio/internal/tasks"

	"github.com/google/wire"
)

// Wire Set
var TaskSet = wire.NewSet(
	tasks.NewFactTask,
)
