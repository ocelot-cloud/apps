package mocks

import "github.com/stretchr/testify/mock"

type WaiterMock struct {
	mock.Mock
}

func (m *WaiterMock) WaitPort(port string) error {
	args := m.Called(port)
	return args.Error(0)
}

func (m *WaiterMock) WaitWeb(url string) error {
	args := m.Called(url)
	return args.Error(0)
}
