package ingestion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// mux is the HTTP request multiplexer used with the test server
	mux *http.ServeMux

	// server is a test HTTP server used to provide mock API responses
	server *httptest.Server

	// client is the Recurly client being tested
	client *Client
)

// setup sets up a test HTTP server along with a ingestion.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient("access_token", nil)
	client.baseURL = server.URL + "/"
}

func teardown() {
	server.Close()
}

func areEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	err := json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

func jsonStringAsInterface(s string) (interface{}, error) {
	var o1 interface{}
	err := json.Unmarshal([]byte(s), &o1)
	return o1, err
}

func Test_interfaceToCSV(t *testing.T) {
	type args struct {
		a interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "[]int64",
			args: args{
				a: []int64{2, 3, 4, 5},
			},
			want: "2,3,4,5",
		},
		{
			name: "[]string",
			args: args{
				a: []string{"candidates.create", "candidates.view", "jobs.view"},
			},
			want: "candidates.create,candidates.view,jobs.view",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, interfaceToCSV(tt.args.a))
		})
	}
}

func Test_spaceDelimit(t *testing.T) {
	type args struct {
		a interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "[]int64",
			args: args{
				a: []int64{2, 3, 4, 5},
			},
			want: "2 3 4 5",
		},
		{
			name: "[]string",
			args: args{
				a: []string{"candidates.create", "candidates.view", "jobs.view"},
			},
			want: "candidates.create candidates.view jobs.view",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, spaceDelimit(tt.args.a))
		})
	}
}
func Test_do_client_error(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v1/client-error", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(400)
		io.WriteString(w, `
		{
			"errors": [
				{
					"message": "Your request included invalid JSON.",
					"field": "email"
				}
			]
		}
		`)
	})

	test := struct {
		wantErr       bool
		wantErrorType error
		clientError   ClientError
	}{
		wantErr:       true,
		wantErrorType: ClientError{},
		clientError: ClientError{
			StatusCode: 400,
			Errors: []Error{
				Error{
					Message: "Your request included invalid JSON.",
					Field:   "email",
				},
			},
		},
	}

	var result interface{}
	req, err := client.newRequest("GET", "client-error", nil, nil)
	if err != nil {
		return
	}
	err = client.do(req, &result)
	switch test.wantErr {
	case true:
		assert.Error(t, err)
		assert.IsType(t, test.wantErrorType, err)
	case false:
		assert.NoError(t, err)
	}

	clientError, _ := IsClientError(err)
	assert.Equal(t, test.clientError, clientError)
}

func Test_do_server_error(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v1/server-error", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(500)
		io.WriteString(w, `
		{
			"errors": [
				{
					"message": "Your request included invalid JSON.",
					"field": "email"
				}
			]
		}
		`)
	})

	test := struct {
		wantErr       bool
		wantErrorType error
		serverError   ServerError
	}{
		wantErr:       true,
		wantErrorType: ServerError{},
		serverError: ServerError{
			StatusCode: 500,
			Errors: []Error{
				Error{
					Message: "Your request included invalid JSON.",
					Field:   "email",
				},
			},
		},
	}

	var result interface{}
	req, err := client.newRequest("GET", "server-error", nil, nil)
	if err != nil {
		return
	}
	err = client.do(req, &result)
	switch test.wantErr {
	case true:
		assert.Error(t, err)
		assert.IsType(t, test.wantErrorType, err)
	case false:
		assert.NoError(t, err)
	}

	serverError, _ := IsServerError(err)
	assert.Equal(t, test.serverError, serverError)
}

func TestClient_newRequest_header_OAuth(t *testing.T) {
	client = NewClient("12345", nil)
	req, err := client.newRequest("GET", "/", nil, nil)
	assert.NoError(t, err)
	authHeader := req.Header.Get("Authorization")
	assert.Equal(t, "Bearer 12345", authHeader)
}

func TestClient_newRequest_header_BasicAuth(t *testing.T) {
	client = NewClientBasicAuth("apiKey", "john.smith@example.com", nil)
	req, err := client.newRequest("GET", "/", nil, nil)
	assert.NoError(t, err)
	authHeader := req.Header.Get("Authorization")
	assert.Equal(t, "Basic YXBpS2V5:", authHeader)
	onBehalfOfHeader := req.Header.Get("On-Behalf-Of")
	assert.Equal(t, "john.smith@example.com", onBehalfOfHeader)
}
