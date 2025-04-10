package queueclient

import (
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/mock"
)

type MockQueueClient struct {
	mock.Mock
}

func (m *MockQueueClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(task, opts)
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}
