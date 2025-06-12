package mocks

import "github.com/stretchr/testify/mock"

type RunnerMock struct {
	mock.Mock
}

func (m *RunnerMock) Run(dir, command string) error {
	args := m.Called(dir, command)
	return args.Error(0)
}
