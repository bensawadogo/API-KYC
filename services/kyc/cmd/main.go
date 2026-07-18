package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/handler"
	"github.com/datakeys/kyc-service/internal/job"
	"github.com/datakeys/kyc-service/internal/middleware"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/datakeys/kyc-service/internal/observability"
	"github.com/datakeys/kyc-service/internal/provider"
	"github.com/datakeys/kyc-service/internal/registry"
	"github.com/datakeys/kyc-service/internal/repository"
	"github.com/datakeys/kyc-service/internal/resilience"
	"github.com/datakeys/kyc-service/internal/service"
	"github.com/datakeys/kyc-service/internal/storage"
	"github.com/datakeys/kyc-service/internal/webhook"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("init logger: %v", err))
	}
	defer zapLogger.Sync()

	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("load config", zap.Error(err))
	}

	ctx := context.Background()
	repo, err := repository.NewPostgresRepository(ctx, cfg.Database.PostgresURL)
	if err != nil {
		zapLogger.Fatal("init repository", zap.Error(err))
	}
	defer repo.Close()

	if err := repository.RunMigrations(repo.Pool()); err != nil {
		zapLogger.Fatal("run migrations", zap.Error(err))
	}
	zapLogger.Info("migrations applied successfully")

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		zapLogger.Warn("redis not available, rate limiting disabled", zap.Error(err))
		redisClient = nil
	}

	baseURL := fmt.Sprintf("http://localhost:%s", cfg.Server.Port)
	var docStorage internal.DocumentStorage
	switch cfg.Storage.Provider {
	case "minio":
		ms, err := storage.NewMinIOStorage(
			cfg.Storage.Endpoint,
			cfg.Storage.AccessKey,
			cfg.Storage.SecretKey,
			cfg.Storage.Bucket,
			cfg.Storage.UseSSL,
			zapLogger,
		)
		if err != nil {
			zapLogger.Fatal("init minio storage", zap.Error(err))
		}
		docStorage = ms
		zapLogger.Info("using minio storage", zap.String("endpoint", cfg.Storage.Endpoint), zap.String("bucket", cfg.Storage.Bucket))
	default:
		docStorage = storage.NewMemoryStorage(baseURL)
		zapLogger.Info("using memory storage")
	}

	smileProvider := provider.NewSmileIDProvider(cfg.Providers.SmileID)
	youverifyProvider := provider.NewYouverifyProvider(cfg.Providers.Youverify)
	sumsubProvider := provider.NewSumSubProvider(cfg.Providers.SumSub)
	localProvider := provider.NewLocalProviderWithSandbox(cfg.Server.Env != "production")

	retryCfg := resilience.RetryConfig{
		MaxAttempts: getEnvInt("PROVIDER_RETRY_MAX", 3),
		BaseDelay:   getDuration("PROVIDER_RETRY_BASE_DELAY", 500*time.Millisecond),
		MaxDelay:    getDuration("PROVIDER_RETRY_MAX_DELAY", 10*time.Second),
		Multiplier:  getEnvFloat("PROVIDER_RETRY_MULTIPLIER", 2.0),
	}
	cbFor := func(name string) resilience.CBConfig {
		return resilience.CBConfig{
			Name:        name,
			MaxFailures: uint32(getEnvInt("CB_MAX_FAILURES", 5)),
			Timeout:     getDuration("CB_TIMEOUT", 30*time.Second),
			MaxRequests: uint32(getEnvInt("CB_MAX_REQUESTS_HALF_OPEN", 2)),
		}
	}

	resilientSmile := resilience.NewResilientProvider(smileProvider, cbFor("smileid"), retryCfg, zapLogger)
	resilientYouverify := resilience.NewResilientProvider(youverifyProvider, cbFor("youverify"), retryCfg, zapLogger)
	resilientSumsub := resilience.NewResilientProvider(sumsubProvider, cbFor("sumsub"), retryCfg, zapLogger)
	resilientLocal := resilience.NewResilientProvider(localProvider, cbFor("local"), retryCfg, zapLogger)

	router := resilience.NewFallbackRouter(
		[]*resilience.ResilientProvider{
			resilientSmile,
			resilientYouverify,
			resilientSumsub,
			resilientLocal,
		},
		zapLogger,
	)

	dlq := resilience.NewDLQ(redisClient, zapLogger)

	tracingShutdown, err := observability.InitTracing(observability.TracingConfig{
		Enabled:     cfg.Observability.TracingEnabled,
		ServiceName: "kyc-service",
		Environment: cfg.Server.Env,
	})
	if err != nil {
		zapLogger.Fatal("init tracing", zap.Error(err))
	}
	defer tracingShutdown()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			size, err := dlq.Len(context.Background())
			if err != nil {
				zapLogger.Debug("dlq size read error", zap.Error(err))
				continue
			}
			observability.DLQSize.Set(float64(size))
		}
	}()

	var amlProvider internal.AMLChecker
	switch cfg.AML.Provider {
	case "opensanctions":
		amlProvider = provider.NewOpenSanctionsProvider(
			cfg.AML.APIKey, cfg.AML.BaseURL, cfg.AML.Threshold, zapLogger,
		)
	default:
		amlProvider = provider.NewLocalAMLProvider()
	}

	kycService := service.NewKYCService(
		repo,
		docStorage,
		nil,
		webhook.NewHTTPSender(),
		registry.New(),
		cfg,
		zapLogger,
		amlProvider,
		repo,
		router,
		dlq,
	)

	cleanupJob := job.NewCleanupJob(
		repo,
		docStorage,
		repo.Pool(),
		redisClient,
		cfg.Compliance.RetentionDays,
		zapLogger,
	)
	cleanupJob.Start()
	defer cleanupJob.Stop()

	kycHandler := handler.NewKYCHandler(
		kycService,
		zapLogger,
		cfg.Providers.SmileID.ApiKey,
		cfg.Providers.SumSub.SecretKey,
		redisClient,
	)
	healthHandler := handler.NewHealthHandler(repo.Pool(), redisClient, router, dlq)

	app := fiber.New(fiber.Config{
		AppName:      "kyc-service",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    11 * 1024 * 1024,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success":   false,
				"error":     err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.MetricsAndTracing())

	app.Use(func(c fiber.Ctx) error {
		c.Response().Header.Set("X-Content-Type-Options", "nosniff")
		c.Response().Header.Set("X-Frame-Options", "DENY")
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Response().Header.Set("X-Request-ID", reqID)
		return c.Next()
	})

	app.Get("/health/live", healthHandler.Live)
	app.Get("/health/ready", healthHandler.Ready)

	app.Get("/docs/openapi.yaml", func(c fiber.Ctx) error {
		return c.SendFile("./api/openapi.yaml")
	})
	app.Get("/docs", func(c fiber.Ctx) error {
		swaggerUI := `<!DOCTYPE html>
<html>
<head><title>DATAKEYS KYC API</title>
<meta charset="utf-8"/>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
SwaggerUIBundle({url:"/docs/openapi.yaml",dom_id:"#swagger-ui",
presets:[SwaggerUIBundle.presets.apis,SwaggerUIBundle.SwaggerUIStandalonePreset]});
</script>
</body></html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerUI)
	})

	app.Get("/metrics", func(c fiber.Ctx) error {
		h := fasthttpadaptor.NewFastHTTPHandler(
			promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
				ErrorHandling: promhttp.ContinueOnError,
			}),
		)
		h(c.RequestCtx())
		return nil
	})

	idmpStore := middleware.NewIdempotencyStore(redisClient, zapLogger)
	smileAllowlist := middleware.NewIPAllowlist(cfg.Webhooks.SmileIDAllowedCIDRs, zapLogger)
	sumsubAllowlist := middleware.NewIPAllowlist(cfg.Webhooks.SumSubAllowedCIDRs, zapLogger)

	v1 := app.Group("/v1")
	kyc := v1.Group("/kyc")

	// Routes publiques (pas d'auth)
	kyc.Get("/countries", kycHandler.ListCountries)
	kyc.Get("/countries/:code/doctypes", kycHandler.ListDocTypes)

	// /v1/kyc/initiate — rate limit + idempotence + auth scope
	initiateGroup := kyc.Group("/initiate")
	initiateGroup.Use(middleware.RequireAPIKey(repo, model.ScopeKYCInitiate))
	initiateGroup.Use(middleware.RateLimitPerKey(redisClient))
	initiateGroup.Use(idmpStore.Middleware)
	initiateGroup.Post("", kycHandler.Initiate)

	// /v1/kyc/status/:verification_id — rate limit + auth scope
	statusGroup := kyc.Group("/status")
	statusGroup.Use(middleware.RequireAPIKey(repo, model.ScopeKYCStatus))
	statusGroup.Use(middleware.RateLimitPerKey(redisClient))
	statusGroup.Get("/:verification_id", kycHandler.GetStatus)

	// Webhooks — IP allowlist (pas d'API key)
	webhookGroup := kyc.Group("/webhook")

	smileGroup := webhookGroup.Group("/smileid")
	smileGroup.Use(smileAllowlist.Middleware)
	smileGroup.Post("", kycHandler.SmileIDWebhook)

	sumsubGroup := webhookGroup.Group("/sumsub")
	sumsubGroup.Use(sumsubAllowlist.Middleware)
	sumsubGroup.Post("", kycHandler.SumSubWebhook)

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port)
		zapLogger.Info("starting kyc service", zap.String("addr", addr), zap.String("env", cfg.Server.Env))
		if err := app.Listen(addr); err != nil {
			zapLogger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down kyc service")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		zapLogger.Error("shutdown error", zap.Error(err))
	}
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}

func getEnvFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return f
}
