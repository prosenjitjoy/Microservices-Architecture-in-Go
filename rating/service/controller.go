package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"main/rating/model"
	"main/rating/repository"
	"main/util"

	"github.com/apache/pulsar-client-go/pulsar"
)

// ErrNotFound is returned when no ratings are found for a record.
var ErrNotFound = errors.New("ratings not found for a record")

type ratingRepository interface {
	Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

// RatingService defines a rating service controller.
type RatingService struct {
	repo ratingRepository
	cfg  *util.ConfigDatabase
}

// New creates a rating service controller.
func New(repo ratingRepository, config *util.ConfigDatabase) *RatingService {
	return &RatingService{
		repo: repo,
		cfg:  config,
	}
}

// GetAggregatedRating returns the aggregated rating for a record or ErrNotFound if there are no ratings for it.
func (s *RatingService) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	ratings, err := s.repo.Get(ctx, recordID, recordType)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}
	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	return sum / float64(len(ratings)), nil
}

// PutRating writes a rating for a given record
func (s *RatingService) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return s.repo.Put(ctx, recordID, recordType, rating)
}

// StartConsume starts consuming the rating events.
func (s *RatingService) StartConsume(ctx context.Context) error {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               s.cfg.PulsarURL,
		ConnectionTimeout: s.cfg.ConnectionTimeout,
		OperationTimeout:  s.cfg.OperationTimeout,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	channel := make(chan pulsar.ConsumerMessage, 100)

	options := pulsar.ConsumerOptions{
		Topic:            s.cfg.TopicName,
		SubscriptionName: s.cfg.SubscriberName,
		Type:             pulsar.Exclusive,
		MessageChannel:   channel,
	}

	consumer, err := client.Subscribe(options)
	if err != nil {
		return err
	}
	defer consumer.Close()

	for cm := range channel {
		consumer := cm.Consumer
		msg := cm.Message
		fmt.Printf("Consumer %s received a message, msgId: %v, content: %s\n", consumer.Name(), msg.ID(), string(msg.Payload()))

		var event model.RatingEvent
		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			return err
		}

		if err := s.PutRating(ctx, event.RecordID, event.RecordType, &model.Rating{
			UserID: event.UserID,
			Value:  event.Value,
		}); err != nil {
			return err
		}

		consumer.Ack(msg)
	}

	return nil
}
