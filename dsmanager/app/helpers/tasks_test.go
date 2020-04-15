package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	manager := NewDefaultTaskManager()
	task := manager.NewTask("title")
	require.Equal(t, "title", task.GetData().Description)
}

func TestGetData(t *testing.T) {
	manager := NewDefaultTaskManager()
	task := manager.NewTask("title")
	task2, ok := task.(*Task)
	require.True(t, ok)
	require.Equal(t, "title", task2.Data.Description)

	task.GetData().Description = "new desc"
	require.Equal(t, "new desc", task2.Data.Description)
}
