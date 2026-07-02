package queue

import (
	"context"
	"log"
	"strconv"
	"sync"

	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/observability"
	"disaster_alert_backend/internal/services"
)

type NotificationQueue struct {
	jobs       chan jobs.NotificationJob
	workerSize int
	service    *services.FirebaseMessagingService
	wg         sync.WaitGroup
}

func NewNotificationQueue(
	service *services.FirebaseMessagingService,
	bufferSize int,
	workerSize int,
) *NotificationQueue {
	return &NotificationQueue{
		jobs:       make(chan jobs.NotificationJob, bufferSize),
		workerSize: workerSize,
		service:    service,
	}
}

func (q *NotificationQueue) Start(ctx context.Context) {
	for i := 1; i <= q.workerSize; i++ {
		q.wg.Add(1)

		go q.worker(ctx, i)
	}

	log.Printf("Notification queue started with %d workers\n", q.workerSize)

	observability.Info(ctx, "Notification queue started", observability.Fields{
		"module":       "notification_queue",
		"worker_count": strconv.Itoa(q.workerSize),
	})
}

func (q *NotificationQueue) Dispatch(job jobs.NotificationJob) bool {
	select {
	case q.jobs <- job:
		return true
	default:
		observability.Warn(context.Background(), "Notification queue is full", observability.Fields{
			"module":      "notification_queue",
			"target_type": string(job.TargetType),
			"title":       job.Title,
		})

		return false
	}
}

func (q *NotificationQueue) Stop() {
	close(q.jobs)
	q.wg.Wait()

	log.Println("Notification queue stopped")

	observability.Info(context.Background(), "Notification queue stopped", observability.Fields{
		"module": "notification_queue",
	})
}

func (q *NotificationQueue) worker(ctx context.Context, workerID int) {
	defer q.wg.Done()

	log.Printf("Notification worker %d started\n", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Notification worker %d stopped by context\n", workerID)
			return

		case job, ok := <-q.jobs:
			if !ok {
				log.Printf("Notification worker %d stopped because queue closed\n", workerID)
				return
			}

			q.processJob(ctx, workerID, job)
		}
	}
}

func (q *NotificationQueue) processJob(ctx context.Context, workerID int, job jobs.NotificationJob) {
	var err error

	switch job.TargetType {
	case jobs.TargetToken:
		_, err = q.service.SendToToken(
			job.Token,
			job.Title,
			job.Body,
			job.Data,
		)

	case jobs.TargetUser:
		err = q.service.SendToUser(
			job.UserID,
			job.Title,
			job.Body,
			job.Data,
		)

	case jobs.TargetAll:
		err = q.service.SendToAll(
			job.Title,
			job.Body,
			job.Data,
		)

	default:
		log.Printf("Worker %d received unknown notification target type: %s\n", workerID, job.TargetType)

		observability.Warn(ctx, "Unknown notification target type", observability.Fields{
			"module":      "notification_queue",
			"worker_id":   strconv.Itoa(workerID),
			"target_type": string(job.TargetType),
			"title":       job.Title,
		})

		return
	}

	if err != nil {
		log.Printf("Worker %d failed to send notification: %v\n", workerID, err)

		observability.Error(ctx, "Notification job failed", err, observability.Fields{
			"module":      "notification_queue",
			"worker_id":   strconv.Itoa(workerID),
			"target_type": string(job.TargetType),
			"title":       job.Title,
		})

		return
	}

	log.Printf("Worker %d sent notification successfully: %s\n", workerID, job.Title)
}
