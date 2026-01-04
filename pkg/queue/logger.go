package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/dbunt1tled/fiber-go-api/pkg/f"
	"github.com/hibiken/asynq"
)

func loggingMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		logSuccess(t, fmt.Sprintf("⏱ Started task: type=%s id=%s", t.Type(), t.ResultWriter().TaskID()))
		err := next.ProcessTask(ctx, t)

		if err != nil {
			logError(
				t,
				fmt.Sprintf(
					"✗ Failed task: type=%s id=%s error=%s (%s)",
					t.Type(),
					t.ResultWriter().TaskID(),
					err.Error(),
					f.RuntimeStatistics(start, false),
				),
				err,
			)
		} else {
			logSuccess(t, fmt.Sprintf(
				"✓ Completed task: type=%s id=%s (%s)",
				t.Type(),
				t.ResultWriter().TaskID(),
				f.RuntimeStatistics(start, false)),
			)
		}

		return err
	})
}
