package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateJSONUsesSchemaWithoutTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if _, exists := request["tools"]; exists {
			t.Fatal("formatting stage must not include tools")
		}
		generationConfig, exists := request["generationConfig"].(map[string]any)
		if !exists {
			t.Fatal("formatting stage must request structured output")
		}
		if got := generationConfig["responseMimeType"]; got != "application/json" {
			t.Fatalf("unexpected responseMimeType: %#v", got)
		}
		responseSchema, exists := generationConfig["responseJsonSchema"].(map[string]any)
		if !exists || responseSchema["type"] != "object" {
			t.Fatalf("unexpected responseJsonSchema: %#v", generationConfig["responseJsonSchema"])
		}
		if _, exists := generationConfig["responseFormat"]; exists {
			t.Fatal("generateContent request must not use the incompatible responseFormat field")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"title\":\"測試日報\"}"}]}}]}`))
	}))
	defer server.Close()

	client := testClient(server)
	var output struct {
		Title string `json:"title"`
	}
	if err := client.GenerateJSON(context.Background(), "Format report", map[string]any{"type": "object"}, &output); err != nil {
		t.Fatalf("GenerateJSON returned an error: %v", err)
	}
	if output.Title != "測試日報" {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func testClient(server *httptest.Server) *Client {
	return NewClient(Config{
		APIKey:     "test-key",
		Model:      "test-model",
		BaseURL:    server.URL + "/v1beta",
		HTTPClient: server.Client(),
	})
}
