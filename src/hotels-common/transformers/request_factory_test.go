package transformers

import (
	"bytes"
	"encoding/json"
	"errors"
	"hotels-common/models"
	"io"
	"io/ioutil"
	"net/url"
	"testing"
)

type mockRequestWrapper struct {
	params  map[string]string
	headers map[string][]string
	body    []byte
	method  string
	url     *url.URL
	query   url.Values
	path    string
}

func (m *mockRequestWrapper) Params() map[string]string    { return m.params }
func (m *mockRequestWrapper) Headers() map[string][]string { return m.headers }
func (m *mockRequestWrapper) Body() (r io.ReadCloser) {
	return ioutil.NopCloser(bytes.NewReader(m.body))
}
func (m *mockRequestWrapper) Method() string    { return m.method }
func (m *mockRequestWrapper) URL() *url.URL     { return m.url }
func (m *mockRequestWrapper) Query() url.Values { return m.query }
func (m *mockRequestWrapper) Path() string      { return m.path }

func TestGetRequestFactory_TBO(t *testing.T) {
	factory, err := GetRequestFactory(models.ProviderTBO)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if factory == nil {
		t.Fatal("expected factory, got nil")
	}
}

func TestGetRequestFactory_Unknown(t *testing.T) {
	_, err := GetRequestFactory("unknown")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTBORequestFactory_BuildRequest_Success(t *testing.T) {
	factory := &TBORequestFactory{}
	build := factory.BuildRequest()
	mockReq := &mockRequestWrapper{
		params:  map[string]string{"foo": "bar"},
		headers: map[string][]string{"X-Test": {"1"}},
		body:    []byte(`{}`),
		method:  "GET",
		url:     &url.URL{Scheme: "http", Host: "test.com"},
		query:   url.Values{"checkin": {"2024-01-01"}, "checkout": {"2024-01-02"}},
		path:    "/test",
	}
	handler := build(map[string]interface{}{})
	result, err := handler(mockReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	rw, ok := result.(requestWrapper)
	if !ok {
		t.Fatalf("expected requestWrapper, got %T", result)
	}
	if rw.method != "POST" {
		t.Errorf("expected method POST, got %s", rw.method)
	}
	body, _ := ioutil.ReadAll(rw.body)
	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		t.Errorf("body is not valid JSON: %v", err)
	}
}

func TestTBORequestFactory_BuildRequest_UnknownType(t *testing.T) {
	factory := &TBORequestFactory{}
	build := factory.BuildRequest()
	handler := build(map[string]interface{}{})
	_, err := handler("not a request wrapper")
	if !errors.Is(err, unknownTypeErr) {
		t.Fatalf("expected unknownTypeErr, got %v", err)
	}
}
