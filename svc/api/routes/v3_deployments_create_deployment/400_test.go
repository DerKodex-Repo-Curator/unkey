package handler_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unkeyed/unkey/svc/api/internal/testutil"
	"github.com/unkeyed/unkey/svc/api/openapi"
	handler "github.com/unkeyed/unkey/svc/api/routes/v3_deployments_create_deployment"
)

func TestOCIValidation(t *testing.T) {
	h := testutil.NewHarness(t)
	setup := h.CreateTestDeploymentSetup()
	route := &handler.Handler{DB: h.DB, CtrlClient: &testutil.MockDeploymentClient{}}
	h.Register(route)

	base := map[string]any{
		"project":     setup.Project.Slug,
		"app":         setup.App.Slug,
		"environment": setup.Environment.Slug,
	}

	for _, tc := range []struct {
		name   string
		source map[string]any
	}{
		{name: "missing source", source: map[string]any{}},
		{name: "missing OCI image", source: map[string]any{"oci": map[string]any{}}},
		{name: "whitespace OCI image", source: map[string]any{"oci": map[string]any{"image": "   "}}},
		{name: "multiple sources", source: map[string]any{"oci": map[string]any{"image": "nginx"}, "git": map[string]any{"branch": "main"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := make(map[string]any, len(base)+len(tc.source))
			for key, value := range base {
				body[key] = value
			}
			for key, value := range tc.source {
				body[key] = value
			}

			headers := http.Header{
				"Content-Type":  {"application/json"},
				"Authorization": {"Bearer " + setup.RootKey},
			}
			res := testutil.CallRoute[map[string]any, openapi.BadRequestErrorResponse](h, route, headers, body)
			require.Equal(t, http.StatusBadRequest, res.Status, "received: %s", res.RawBody)
		})
	}
}
