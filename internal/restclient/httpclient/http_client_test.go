package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPClient_Do(t *testing.T) {
	type fields struct {
		cxProfile  HTTPProfile
		ctx        context.Context
		httpClient http.Client
	}
	type args struct {
		baseURL string
		req     *Request
	}
	cxProfile := HTTPProfile{
		Hostname: "host",
		APIRoot:  "api",
	}
	request := Request{
		Method: "GET",
		Body:   map[string]any{},
		Query:  map[string][]string{},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		want1   []byte
		wantErr bool
	}{
		{
			name:    "r is nil",
			fields:  fields{},
			args:    args{},
			want:    -1,
			want1:   nil,
			wantErr: true,
		},
		{
			name:    "Hostname and APIRoot are required",
			fields:  fields{},
			args:    args{req: &request},
			want:    -1,
			want1:   nil,
			wantErr: true,
		},
		{
			name:    "lookup error on host",
			fields:  fields{cxProfile: cxProfile, ctx: context.Background(), httpClient: http.Client{}},
			args:    args{req: &request},
			want:    -1,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "connect: connection refused",
			fields: fields{cxProfile: HTTPProfile{
				APIRoot:  "api",
				Hostname: "localhost",
			}, ctx: context.Background(), httpClient: http.Client{}},
			args:    args{req: &request},
			want:    -1,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cxProfile:  tt.fields.cxProfile,
				ctx:        tt.fields.ctx,
				httpClient: tt.fields.httpClient,
			}
			got, got1, err := c.Do(tt.args.baseURL, tt.args.req)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPClient.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HTTPClient.Do() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("HTTPClient.Do() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestHTTPClientProxyEnvironment(t *testing.T) {
	// net/http caches proxy environment variables process-wide. Run each case
	// in a fresh process so neither other tests nor the host environment leak in.
	if scenario := os.Getenv("ONTAP_PROXY_TEST_CASE"); scenario != "" {
		testProxyEnvironment(t, scenario)
		return
	}
	for _, scenario := range []string{"http", "https", "no_proxy", "direct"} {
		t.Run(scenario, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestHTTPClientProxyEnvironment$")
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				switch strings.ToUpper(key) {
				case "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "REQUEST_METHOD", "ONTAP_PROXY_TEST_CASE":
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "ONTAP_PROXY_TEST_CASE="+scenario)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%s: %v\n%s", scenario, err, output)
			}
		})
	}
}

func testProxyEnvironment(t *testing.T, scenario string) {
	requests := make(chan string, 1)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.Host
		// For HTTPS, observing CONNECT proves the transport reached the proxy
		// without resolving the deliberately invalid target hostname locally.
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxy.Close()

	scheme := "https"
	switch scenario {
	case "http":
		scheme = "http"
		t.Setenv("HTTP_PROXY", proxy.URL)
	case "https":
		t.Setenv("HTTPS_PROXY", proxy.URL)
	case "no_proxy":
		t.Setenv("HTTPS_PROXY", proxy.URL)
		t.Setenv("NO_PROXY", "ontap-proxy-repro.invalid")
	}

	client := NewClient(context.Background(), HTTPProfile{
		Username: "fsxadmin", Password: "test", ValidateCerts: true,
	}, "test")
	client.httpClient.Timeout = 5 * time.Second
	transport := client.httpClient.Transport.(*http.Transport)
	defer transport.CloseIdleConnections()
	req, err := http.NewRequest(http.MethodGet, scheme+"://ontap-proxy-repro.invalid/api/cluster", nil)
	if err != nil {
		t.Fatal(err)
	}
	if scenario == "no_proxy" || scenario == "direct" {
		if transport.Proxy != nil {
			selected, err := transport.Proxy(req)
			if err != nil || selected != nil {
				t.Fatalf("expected direct connection, got proxy %v, error %v", selected, err)
			}
		}
		return
	}

	response, requestErr := client.httpClient.Do(req)
	if response != nil {
		response.Body.Close()
	}
	expected := "GET ontap-proxy-repro.invalid"
	if scenario == "https" {
		expected = "CONNECT ontap-proxy-repro.invalid:443"
	}
	select {
	case got := <-requests:
		if got != expected {
			t.Fatalf("proxy received %q, want %q", got, expected)
		}
	default:
		t.Fatalf("request did not reach proxy: %v", requestErr)
	}
}
