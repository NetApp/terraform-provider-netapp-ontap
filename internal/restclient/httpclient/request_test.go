package httpclient

import (
	"context"
	"net/url"
	"testing"
)

func TestRequest_BuildURL(t *testing.T) {
	cxProfile := HTTPProfile{
		Hostname: "host",
		APIRoot:  "api",
	}
	client := &HTTPClient{
		cxProfile: cxProfile,
	}
	clientEmptyValues := &HTTPClient{
		cxProfile: HTTPProfile{},
	}
	query := url.Values{}
	query.Add("fields", "f1,f2")
	type fields struct {
		Method string
		Body   map[string]interface{}
		Query  url.Values
	}
	type args struct {
		c       *HTTPClient
		baseURL string
		uuid    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "test1", fields: fields{Method: "GET", Body: nil, Query: nil}, args: args{c: client, baseURL: "cluster"}, want: "https://host/api/cluster", wantErr: false},
		{name: "test2", fields: fields{Method: "GET", Body: nil, Query: query}, args: args{c: client, baseURL: "cluster", uuid: "123"}, want: "https://host/api/cluster/123?fields=f1%2Cf2", wantErr: false},
		{name: "test3", fields: fields{Method: "GET", Body: nil, Query: query}, args: args{c: nil, baseURL: "cluster", uuid: "123"}, want: "", wantErr: true},
		{name: "test4", fields: fields{Method: "GET", Body: nil, Query: query}, args: args{c: clientEmptyValues, baseURL: "cluster", uuid: "123"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{
				Method: tt.fields.Method,
				Body:   tt.fields.Body,
				Query:  tt.fields.Query,
			}
			got, err := r.BuildURL(tt.args.c, tt.args.baseURL, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.BuildURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Request.BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_BuildHTTPReq(t *testing.T) {
	cxProfile := HTTPProfile{
		Hostname: "host",
		APIRoot:  "api",
	}
	client := &HTTPClient{
		cxProfile: cxProfile,
		ctx:       context.TODO(),
	}
	testURL := "https://host/api/cluster"
	testURLQ := "https://host/api/cluster?fields=f1%2Cf2"
	body := make(map[string]any)
	body["fields"] = "f1,f2"
	marshalErrorBody := map[string]interface{}{
		"foo": make(chan int),
	}
	query := url.Values{}
	query.Add("fields", "f1,f2")
	type fields struct {
		Method string
		Body   map[string]interface{}
		Query  url.Values
	}
	type wants struct {
		Method  string
		nilBody bool
		Query   url.Values
		url     string
	}
	type args struct {
		c       *HTTPClient
		baseURL string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    wants
		wantErr bool
	}{
		{name: "test1", fields: fields{Method: "GET", Body: nil, Query: nil}, args: args{c: client, baseURL: "cluster"}, want: wants{Method: "GET", nilBody: true, Query: nil, url: testURL}, wantErr: false},
		{name: "test2", fields: fields{Method: "POST", Body: nil, Query: nil}, args: args{c: client, baseURL: "cluster"}, want: wants{Method: "POST", nilBody: true, Query: nil, url: testURL}, wantErr: false},
		{name: "test3", fields: fields{Method: "GET", Body: nil, Query: query}, args: args{c: client, baseURL: "cluster"}, want: wants{Method: "GET", nilBody: true, Query: nil, url: testURLQ}, wantErr: false},
		{name: "test4", fields: fields{Method: "POST", Body: body, Query: nil}, args: args{c: client, baseURL: "cluster"}, want: wants{Method: "POST", nilBody: false, Query: nil, url: testURL}, wantErr: false},
		{name: "test5", fields: fields{Method: "BAD METHOD", Body: body, Query: nil}, args: args{c: client, baseURL: "cluster"}, want: wants{Method: "BAD", nilBody: false, Query: nil, url: testURL}, wantErr: true},
		{name: "test6", fields: fields{Method: "POST", Body: marshalErrorBody, Query: nil}, args: args{c: client, baseURL: "cluster"}, want: wants{Method: "BAD", nilBody: false, Query: nil, url: testURL}, wantErr: true},
		{name: "test7", fields: fields{Method: "POST", Body: marshalErrorBody, Query: nil}, args: args{c: nil, baseURL: "cluster"}, want: wants{Method: "BAD", nilBody: false, Query: nil, url: testURL}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{
				Method: tt.fields.Method,
				Body:   tt.fields.Body,
				Query:  tt.fields.Query,
			}
			got, err := r.BuildHTTPReq(tt.args.c, tt.args.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.BuildHTTPReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if got != nil {
					t.Errorf("Request.BuildHTTPReq() error = %v, got %v", err, got)
				}
				return
			}
			if tt.want.Method != got.Method {
				t.Errorf("Request.BuildHTTPReq() = %v, want %v", got, tt.want)
			}
			if tt.want.url != got.URL.String() {
				t.Errorf("Request.BuildHTTPReq() = %v, want %v", got.URL.String(), tt.want.url)
			}
			if (got.Body == nil) != tt.want.nilBody {
				t.Errorf("Request.BuildHTTPReq() = %v, want %v", got, tt.want)
			}
		})
	}
}
