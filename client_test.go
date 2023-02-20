package nodion

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern string, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	mux.HandleFunc(pattern, handler)

	return client
}

func readFileHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		authorization := req.Header.Get("Authorization")
		if authorization != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", authorization), http.StatusUnauthorized)
			return
		}

		//nolint:gosec // Only for testing purpose.
		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		rw.WriteHeader(statusCode)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func TestClient_CreateZone(t *testing.T) {
	client := setupTest(t, "/dns_zones", readFileHandler(http.MethodPost, http.StatusOK, "create-dns-zone.json"))

	zone, err := client.CreateZone(context.Background(), "xxx")
	require.NoError(t, err)

	require.NotNil(t, zone)

	// hack to compare date
	location := zone.CreatedAt.Location()

	expected := &Zone{
		ID:        "52be5f1b-fee7-4a42-b668-85890c41be5b",
		Name:      "nodionsample.com",
		CreatedAt: time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		UpdatedAt: time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		Records: []Record{
			{
				ID:         "5ed9465f-f8c6-432d-9474-3e9880f6adfe",
				RecordType: "a",
				Name:       "@",
				Content:    "1.2.3.4",
				TTL:        3600,
				ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
				CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			},
			{
				ID:         "60a0647b-0b08-4dc0-8d51-4e21c799457c",
				RecordType: "a",
				Name:       "*",
				Content:    "1.2.3.4",
				TTL:        3600,
				ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
				CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			},
			{
				ID:         "b4748041-f3b2-40f3-9217-9af016c5937f",
				RecordType: "a",
				Name:       "www",
				Content:    "1.2.3.4",
				TTL:        3600,
				ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
				CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			},
			{
				ID:         "d13e85ce-7d04-4770-9197-19f87f35e6a8",
				RecordType: "ns",
				Name:       "@",
				Content:    "ns1.nodion.com",
				TTL:        3600,
				ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
				CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			},
			{
				ID:         "f5454bf7-f89b-45a4-981f-7783102fd389",
				RecordType: "ns",
				Name:       "@",
				Content:    "ns2.nodion.com",
				TTL:        3600,
				ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
				CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_CreateZone_error(t *testing.T) {
	client := setupTest(t, "/dns_zones", readFileHandler(http.MethodPost, http.StatusBadRequest, "create-dns-zone-error.json"))

	_, err := client.CreateZone(context.Background(), "")
	require.Error(t, err)
}

func TestClient_DeleteZone(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx", readFileHandler(http.MethodDelete, http.StatusOK, "delete-dns-zone.json"))

	result, err := client.DeleteZone(context.Background(), "xxx")
	require.NoError(t, err)

	assert.True(t, result)
}

func TestClient_DeleteZone_error(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx", readFileHandler(http.MethodDelete, http.StatusNotFound, "delete-dns-zone-error.json"))

	result, err := client.DeleteZone(context.Background(), "xxx")
	require.Error(t, err)

	assert.False(t, result)
}

func TestClient_GetZones(t *testing.T) {
	client := setupTest(t, "/dns_zones", readFileHandler(http.MethodGet, http.StatusOK, "get-dns-zones.json"))

	zones, err := client.GetZones(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, zones, 1)

	// hack to compare date
	location := zones[0].CreatedAt.Location()

	expected := []Zone{
		{
			ID:        "52be5f1b-fee7-4a42-b668-85890c41be5b",
			Name:      "nodionsample.com",
			CreatedAt: time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			UpdatedAt: time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			Records: []Record{
				{
					ID:         "5ed9465f-f8c6-432d-9474-3e9880f6adfe",
					RecordType: "a",
					Name:       "@",
					Content:    "1.2.3.4",
					TTL:        3600,
					ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
					CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
					UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				},
				{
					ID:         "60a0647b-0b08-4dc0-8d51-4e21c799457c",
					RecordType: "a",
					Name:       "*",
					Content:    "1.2.3.4",
					TTL:        3600,
					ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
					CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
					UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				},
				{
					ID:         "b4748041-f3b2-40f3-9217-9af016c5937f",
					RecordType: "a",
					Name:       "www",
					Content:    "1.2.3.4",
					TTL:        3600,
					ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
					CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
					UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				},
				{
					ID:         "d13e85ce-7d04-4770-9197-19f87f35e6a8",
					RecordType: "ns",
					Name:       "@",
					Content:    "ns1.nodion.com",
					TTL:        3600,
					ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
					CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
					UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				},
				{
					ID:         "f5454bf7-f89b-45a4-981f-7783102fd389",
					RecordType: "ns",
					Name:       "@",
					Content:    "ns2.nodion.com",
					TTL:        3600,
					ZoneID:     "52be5f1b-fee7-4a42-b668-85890c41be5b",
					CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
					UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
				},
			},
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZones_error(t *testing.T) {
	client := setupTest(t, "/dns_zones", readFileHandler(http.MethodGet, http.StatusNotFound, "get-dns-zones-error.json"))

	_, err := client.GetZones(context.Background(), nil)
	require.Error(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx/records", readFileHandler(http.MethodGet, http.StatusOK, "get-dns-zones-records.json"))

	records, err := client.GetRecords(context.Background(), "xxx", nil)
	require.NoError(t, err)

	require.Len(t, records, 5)

	// hack to compare date
	location := records[0].CreatedAt.Location()

	expected := []Record{
		{
			ID:         "8231bac6-39f0-4f06-bd6c-076fb9abea9e",
			RecordType: "a",
			Name:       "@",
			Content:    "1.2.3.4",
			TTL:        3600,
			ZoneID:     "",
			CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		},
		{
			ID:         "25adc6de-ee1e-4e94-916a-be3f4bcaa586",
			RecordType: "a",
			Name:       "*",
			Content:    "1.2.3.4",
			TTL:        3600,
			ZoneID:     "",
			CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		},
		{
			ID:         "843fa60c-dc30-47c4-a818-fee31118a43f",
			RecordType: "a",
			Name:       "www",
			Content:    "1.2.3.4",
			TTL:        3600,
			ZoneID:     "",
			CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		},
		{
			ID:         "a10acb05-c76f-4170-9e27-74bb9a6c6cdc",
			RecordType: "ns",
			Name:       "@",
			Content:    "ns1.nodion.com",
			TTL:        3600,
			ZoneID:     "",
			CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		},
		{
			ID:         "924f32d4-b10f-47ef-a293-adbc7169e885",
			RecordType: "ns",
			Name:       "@",
			Content:    "ns2.nodion.com",
			TTL:        3600,
			ZoneID:     "",
			CreatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
			UpdatedAt:  time.Date(2023, time.January, 1, 10, 0, 0, 0, location),
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := setupTest(t, "/dns_zones/records", readFileHandler(http.MethodGet, http.StatusNotFound, "get-dns-zones-records-error.json"))

	_, err := client.GetRecords(context.Background(), "", nil)
	require.Error(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx/records", readFileHandler(http.MethodPost, http.StatusOK, "create-dns-zone-record.json"))

	record := Record{
		RecordType: TypeA,
		Name:       "www",
		Content:    "1.2.3.4",
		TTL:        60,
	}

	newRecord, err := client.CreateRecord(context.Background(), "xxx", record)
	require.NoError(t, err)

	require.NotNil(t, newRecord)

	// hack to compare date
	location := newRecord.CreatedAt.Location()

	expected := &Record{
		ID:         "748d688a-3004-4b84-b8b8-8cb2e07c5c71",
		RecordType: "a",
		Name:       "www",
		Content:    "1.2.3.4",
		TTL:        60,
		ZoneID:     "",
		CreatedAt:  time.Date(2023, time.February, 10, 21, 32, 54, 749000000, location),
		UpdatedAt:  time.Date(2023, time.February, 10, 21, 32, 54, 749000000, location),
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx/records", readFileHandler(http.MethodPost, http.StatusBadRequest, "create-dns-zone-record-error.json"))

	record := Record{
		RecordType: TypeA,
		Name:       "www",
		Content:    "1.2.3.4",
		TTL:        60,
	}

	_, err := client.CreateRecord(context.Background(), "xxx", record)
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx/records/yyy", readFileHandler(http.MethodDelete, http.StatusOK, "delete-dns-zone-record.json"))

	result, err := client.DeleteRecord(context.Background(), "xxx", "yyy")
	require.NoError(t, err)

	assert.True(t, result)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "/dns_zones/xxx/records/yyy", readFileHandler(http.MethodDelete, http.StatusNotFound, "delete-dns-zone-record-error.json"))

	result, err := client.DeleteRecord(context.Background(), "xxx", "yyy")
	require.Error(t, err)

	assert.False(t, result)
}
