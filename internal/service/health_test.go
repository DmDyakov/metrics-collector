package service

import (
	"context"
	"errors"
	"testing"

	"metrics-collector/internal/service/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHealthService_Ping(t *testing.T) {
	t.Run("positive: ping success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := NewHealthService(mockRepo)

		mockRepo.EXPECT().Ping(gomock.Any()).Return(nil)

		err := svc.Ping(context.Background())
		require.NoError(t, err)
	})

	t.Run("negative: ping fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockHealthRepository(ctrl)
		svc := NewHealthService(mockRepo)

		mockRepo.EXPECT().Ping(gomock.Any()).Return(errors.New("db error"))

		err := svc.Ping(context.Background())
		require.Error(t, err)
	})
}
