package handler

import (
	"context"
	"net/http"

	"github.com/unkeyed/unkey/gen/rpc/ctrl"
	"github.com/unkeyed/unkey/pkg/db"
	"github.com/unkeyed/unkey/pkg/zen"
	"github.com/unkeyed/unkey/svc/api/openapi"
	v2 "github.com/unkeyed/unkey/svc/api/routes/v2_deployments_create_deployment"
)

type (
	Request  = openapi.V3DeploymentsCreateDeploymentRequestBody
	Response = openapi.V3DeploymentsCreateDeploymentResponseBody
)

type Handler struct {
	DB         db.Database
	CtrlClient ctrl.DeployServiceClient
}

func (h *Handler) Path() string {
	return "/v3/deployments.createDeployment"
}

func (h *Handler) Method() string {
	return "POST"
}

func (h *Handler) Handle(ctx context.Context, s *zen.Session) error {
	req, err := zen.BindBody[Request](s)
	if err != nil {
		return err
	}

	createReq := v2.CreateDeploymentRequest{
		Project:     req.Project,
		App:         req.App,
		Environment: req.Environment,
		Git:         req.Git,
		OCIImage:    "",
		Deployment:  req.Deployment,
	}
	if req.Oci != nil {
		createReq.OCIImage = req.Oci.Image
	}

	core := v2.Handler{DB: h.DB, CtrlClient: h.CtrlClient}
	deploymentID, err := core.Create(ctx, s, createReq)
	if err != nil {
		return err
	}

	return s.JSON(http.StatusCreated, Response{
		Meta: openapi.Meta{RequestId: s.RequestID()},
		Data: openapi.V3DeploymentsCreateDeploymentResponseData{DeploymentId: deploymentID},
	})
}
