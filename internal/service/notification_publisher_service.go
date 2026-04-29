package service

import (
	// "context"
	"pfo-vector/internal/repository"

	"github.com/redis/go-redis/v9"
)

type NotificationPublisherService struct {
	rdb *redis.Client
	queries *repository.Queries
	batchSize int
	streamName string
}

func NewNotificationPublisherService(rdb *redis.Client,queries *repository.Queries) *NotificationPublisherService{
	return &NotificationPublisherService{
		rdb:rdb,
		queries: queries,
		batchSize: 10,
		streamName: "notifications:stream",
	}
}


