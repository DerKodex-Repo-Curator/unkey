package handler_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ctrlv1 "github.com/unkeyed/unkey/gen/proto/ctrl/v1"
	"github.com/unkeyed/unkey/pkg/db"
	"github.com/unkeyed/unkey/pkg/uid"
	"github.com/unkeyed/unkey/svc/api/internal/testutil"
	"github.com/unkeyed/unkey/svc/api/openapi"
	handler "github.com/unkeyed/unkey/svc/api/routes/v3_deployments_create_deployment"
)

func TestCreateOCIDeployment(t *testing.T) {
	h := testutil.NewHarness(t)
	setup := h.CreateTestDeploymentSetup(testutil.CreateTestDeploymentSetupOptions{
		Permissions: []string{"environment.*.create_deployment"},
	})
	seedDeployableRegion(t, h, setup)

	var captured *ctrlv1.CreateDeploymentRequest
	route := &handler.Handler{
		DB: h.DB,
		CtrlClient: &testutil.MockDeploymentClient{
			CreateDeploymentFunc: func(_ context.Context, req *ctrlv1.CreateDeploymentRequest) (*ctrlv1.CreateDeploymentResponse, error) {
				captured = req
				return &ctrlv1.CreateDeploymentResponse{DeploymentId: "d_test_generated"}, nil
			},
		},
	}
	h.Register(route)

	req := handler.Request{
		Project:     setup.Project.Slug,
		App:         setup.App.Slug,
		Environment: setup.Environment.Slug,
		Oci:         &openapi.DeploymentSourceOCI{Image: "nginx:latest"},
	}
	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + setup.RootKey},
	}
	res := testutil.CallRoute[handler.Request, handler.Response](h, route, headers, req)
	require.Equal(t, http.StatusCreated, res.Status, "received: %s", res.RawBody)
	require.Equal(t, "d_test_generated", res.Body.Data.DeploymentId)
	require.NotNil(t, captured)
	require.Equal(t, "nginx:latest", captured.DockerImage)
}

func seedDeployableRegion(t *testing.T, h *testutil.Harness, setup testutil.DeploymentTestSetup) {
	t.Helper()
	ctx := context.Background()
	regionID := uid.New(uid.RegionPrefix)
	require.NoError(t, db.Query.UpsertRegion(ctx, h.DB.RW(), db.UpsertRegionParams{
		ID:       regionID,
		Name:     uid.New(uid.RegionPrefix),
		Platform: "test",
	}))
	require.NoError(t, db.Query.UpsertAppRegionalSettings(ctx, h.DB.RW(), db.UpsertAppRegionalSettingsParams{
		WorkspaceID:   setup.Workspace.ID,
		AppID:         setup.App.ID,
		EnvironmentID: setup.Environment.ID,
		RegionID:      regionID,
		Replicas:      1,
		CreatedAt:     time.Now().UnixMilli(),
		UpdatedAt:     sql.NullInt64{Valid: false},
	}))
}
