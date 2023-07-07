package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
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
