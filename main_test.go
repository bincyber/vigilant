package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHealthEndpoint(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	server := http.NewServeMux()
	server.Handle("/health", http.HandlerFunc(healthEndpoint))

	server.ServeHTTP(resp, req)

	t.Run("/health returns HTTP 200", func(t *testing.T) {
		assert.Equal(t, resp.Code, http.StatusOK)
	})

	t.Run("/health returns correct response body", func(t *testing.T) {
		assert.Equal(t, resp.Body.String(), "ok")
	})
}

func TestSyncEndpoint(t *testing.T) {

	var testWebHookResponse WebHookResponse

	namespace := "test"

	testWebHookRequest := WebHookRequest{api.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}}

	body, _ := json.Marshal(testWebHookRequest)

	req, _ := http.NewRequest(http.MethodPost, "/sync", bytes.NewBuffer(body))

	resp := httptest.NewRecorder()

	server := http.NewServeMux()
	server.Handle("/sync", http.HandlerFunc(syncEndpoint))

	server.ServeHTTP(resp, req)

	t.Run("/sync returns HTTP 200", func(t *testing.T) {
		assert.Equal(t, resp.Code, http.StatusOK)
	})

	t.Run("/sync returns JSON", func(t *testing.T) {
		err := json.NewDecoder(resp.Body).Decode(&testWebHookResponse)
		assert.Nil(t, err)
	})

	t.Run("/sync returns correct label in response body", func(t *testing.T) {
		labels := map[string]string{
			"name": namespace,
		}

		assert.Equal(t, labels, testWebHookResponse.Labels)
	})

	t.Run("/sync returns NetworkPolicy attachment in response body", func(t *testing.T) {
		assert.Equal(t, len(testWebHookResponse.Attachments), 1)
		assert.Equal(t, testWebHookResponse.Attachments[0].APIVersion, "networking.k8s.io/v1")
		assert.Equal(t, testWebHookResponse.Attachments[0].Kind, "NetworkPolicy")
		assert.Equal(t, testWebHookResponse.Attachments[0].Name, "default-deny-all")
		assert.Equal(t, testWebHookResponse.Attachments[0].Namespace, namespace)
	})
}
