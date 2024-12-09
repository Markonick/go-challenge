package container

import (
	"context"
	"os"
	"strconv"

	"github.com/markonick/gigs-challenge/internal/controllers"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/services"
	"github.com/markonick/gigs-challenge/internal/svix"
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

	must(container.Provide(func(maxWorkers int) *worker.Pool {
		return worker.NewPool(maxWorkers)
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
