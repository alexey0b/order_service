package monitoring_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order_service/internal/infrastructure/monitoring"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

// var (
// 	promoC testcontainers.Container
// )

// func TestMain(m *testing.M) {
// 	ctx := context.Background()

// 	buildContext, err := filepath.Abs("./testdata")
// 	if err != nil {
// 		log.Fatalf("Failed to resolve absolute path: %v\n", err)
// 	}

// 	req := testcontainers.ContainerRequest{
// 		FromDockerfile: testcontainers.FromDockerfile{
// 			Context: buildContext,
// 		},
// 		WaitingFor: wait.ForListeningPort("9090/tcp"),
// 	}

// 	promoC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	code := m.Run()

// 	testcontainers.TerminateContainer(promoC)
// 	os.Exit(code)
// }

func TestPrometheusMetrics(t *testing.T) {
	metrics, err := monitoring.NewPrometheusMetrics()
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer metrics.ObserveRequest(start)
		metrics.IncRequest()

		// Имитируем работу
		time.Sleep(10 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	for i := 0; i < 5; i++ {
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}

	metricsHandler := promhttp.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	metricsHandler.ServeHTTP(w, req)

	bodyStr := w.Body.String()
	require.Contains(t, bodyStr, "app_requests_total 5")
	require.Contains(t, bodyStr, "app_request_duration_seconds_count 5")
	require.Regexp(t, `app_request_duration_seconds_sum 0\.0[5-6][0-9]*`, bodyStr)
}
