package container

import (
	"context"
	"os"
	"strconv"

	"github.com/markonick/gigs-challenge/internal/controllers"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/services"
	"github.com/markonick/gigs-challenge/internal/svix"
	webhook "github.com/markonick/gigs-challenge/internal/webhooks"
	"github.com/markonick/gigs-challenge/internal/worker"
	"go.uber.org/dig"
)

func NewContainer() *dig.Container {
	container := dig.New()

	// Register environment-based providers
	must(container.Provide(func() (string, error) {
		token := os.Getenv("SVIX_AUTH_TOKEN")
		if token == "" {
			logger.Log.Fatal().Msg("SVIX_AUTH_TOKEN is not set")
		}
		return token, nil
	}))

	must(container.Provide(func() (int, error) {
		workers, err := strconv.Atoi(os.Getenv("MAX_WORKERS"))
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("MAX_WORKERS is not set up correctly")
		}
		return workers, nil
	}))

	// Register core services
	must(container.Provide(func(token string) svix.Client {
		return svix.NewClient(token)
	}))

	must(container.Provide(func(client svix.Client) (map[string]string, error) {
		projects := []string{"dev"}
		return svix.InitializeApplications(context.Background(), client, projects)
	}))

	// Add the new TaskProcessor setup
	must(container.Provide(func() *worker.TaskProcessor {
		numWorkers := 5 // Example: set the number of workers
		return worker.NewTaskProcessor(numWorkers)
	}))

	// Register task creation function
	must(container.Provide(func(svixClient svix.Client, projectAppIDs map[string]string) func(models.BaseEvent) worker.Task {
		return func(event models.BaseEvent) worker.Task {
			return webhook.NewWebhookTask(event, svixClient, projectAppIDs)
		}
	}))

	must(container.Provide(services.NewTaskService))
	must(container.Provide(controllers.NewNotificationController))

	return container
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
