package svix

import (
	"context"
	"fmt"

	"github.com/markonick/gigs-challenge/internal/logger"
)

func InitializeApplications(ctx context.Context, client Client, projects []string) (map[string]string, error) {
	projectAppIDs := make(map[string]string)

	for _, projectID := range projects {
		appName := fmt.Sprintf("gigs-webhook-service-%s", projectID)
		appID, err := client.CreateApplication(ctx, appName)
		if err != nil {
			return nil, fmt.Errorf("failed to create application for project %s: %w", projectID, err)
		}

		if err := client.SetupApplicationEndpoints(ctx, appID, projectID); err != nil {
			return nil, fmt.Errorf("failed to setup endpoints for project %s: %w", projectID, err)
		}

		projectAppIDs[projectID] = appID
		logger.Log.Info().
			Str("project", projectID).
			Str("app_id", appID).
			Msg("Successfully set up Svix application and endpoints")
	}

	return projectAppIDs, nil
}
