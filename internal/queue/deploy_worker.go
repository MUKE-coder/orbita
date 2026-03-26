package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type DeployJob struct {
	AppID        uuid.UUID `json:"app_id"`
	DeploymentID uuid.UUID `json:"deployment_id"`
	OrgID        uuid.UUID `json:"org_id"`
	OrgSlug      string    `json:"org_slug"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
}

type DeployWorker struct {
	rdb     *redis.Client
	handler func(ctx context.Context, job DeployJob) error
}

func NewDeployWorker(rdb *redis.Client, handler func(ctx context.Context, job DeployJob) error) *DeployWorker {
	return &DeployWorker{
		rdb:     rdb,
		handler: handler,
	}
}

func (w *DeployWorker) Enqueue(ctx context.Context, job DeployJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("Enqueue: marshal: %w", err)
	}

	queueKey := fmt.Sprintf("orbita:deploy_queue:%s", job.OrgID.String())
	if err := w.rdb.RPush(ctx, queueKey, data).Err(); err != nil {
		return fmt.Errorf("Enqueue: push: %w", err)
	}

	log.Info().
		Str("app_id", job.AppID.String()).
		Str("deployment_id", job.DeploymentID.String()).
		Msg("Deploy job enqueued")

	return nil
}

func (w *DeployWorker) Start(ctx context.Context) {
	log.Info().Msg("Deploy worker started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Deploy worker stopping")
				return
			default:
				// Poll all org queues
				// TODO: real impl would use BLPOP across known org queues
				// For now, this is a simple polling loop
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (w *DeployWorker) ProcessQueue(ctx context.Context, orgID uuid.UUID) error {
	queueKey := fmt.Sprintf("orbita:deploy_queue:%s", orgID.String())

	data, err := w.rdb.LPop(ctx, queueKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("ProcessQueue: pop: %w", err)
	}

	var job DeployJob
	if err := json.Unmarshal(data, &job); err != nil {
		return fmt.Errorf("ProcessQueue: unmarshal: %w", err)
	}

	log.Info().
		Str("app_id", job.AppID.String()).
		Str("deployment_id", job.DeploymentID.String()).
		Msg("Processing deploy job")

	if err := w.handler(ctx, job); err != nil {
		log.Error().Err(err).Str("app_id", job.AppID.String()).Msg("Deploy job failed")
		return err
	}

	return nil
}
