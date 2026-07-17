package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/datakeys/kyc-service/internal/handler"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockService struct {
	initiateFn func(ctx context.Context, req *model.InitiateKYCRequest) (*model.InitiateKYCResponse, error)
	statusFn   func(ctx context.Context, id string) (*model.VerificationResult, error)
}

func (m *mockService) Initiate(ctx context.Context, req *model.InitiateKYCRequest) (*model.InitiateKYCResponse, error) {
	if m.initiateFn != nil {
		return m.initiateFn(ctx, req)
	}
	return &model.InitiateKYCResponse{VerificationID: "v1", Status: "pending", ExpiresIn: 3600, Provider: "smileid"}, nil
}
func (m *mockService) Process(ctx context.Context, verificationID string) error {
	return nil
}
func (m *mockService) GetStatus(ctx context.Context, id string) (*model.VerificationResult, error) {
	if m.statusFn != nil {
		return m.statusFn(ctx, id)
	}
	return &model.VerificationResult{VerificationID: id, Status: "approved"}, nil
}

func setupApp() (*fiber.App, *mockService) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	app := fiber.New()

	kycHandler := handler.NewKYCHandler(svc, logger, "", "", nil)
	app.Post("/v1/kyc/initiate", kycHandler.Initiate)
	app.Get("/v1/kyc/status/:verification_id", kycHandler.GetStatus)
	app.Get("/v1/kyc/countries", kycHandler.ListCountries)
	app.Get("/v1/kyc/countries/:code/doctypes", kycHandler.ListDocTypes)

	return app, svc
}

func TestInitiate_500_ServiceError(t *testing.T) {
	app, svc := setupApp()
	svc.initiateFn = func(ctx context.Context, req *model.InitiateKYCRequest) (*model.InitiateKYCResponse, error) {
		return nil, assert.AnError
	}

	body := `{"phone":"+22670000000","country_code":"BF","doc_type":"NATIONAL_ID","doc_number":"B1234567","full_name":"Test User","consent":true}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/kyc/initiate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetStatus_200_Approved(t *testing.T) {
	app, _ := setupApp()
	req, _ := http.NewRequest(http.MethodGet, "/v1/kyc/status/550e8400-e29b-41d4-a716-446655440000", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestListCountries_200(t *testing.T) {
	app, _ := setupApp()
	req, _ := http.NewRequest(http.MethodGet, "/v1/kyc/countries", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetDocTypes_404_Unknown(t *testing.T) {
	app, _ := setupApp()
	req, _ := http.NewRequest(http.MethodGet, "/v1/kyc/countries/XX/doctypes", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}