package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

func TestConfig_GetConnectionProfile(t *testing.T) {
	type fields struct {
		ConnectionProfiles map[string]ConnectionProfile
		Version            string
	}
	type args struct {
		name string
	}
	cxProfile := ConnectionProfile{}
	cxProfiles := map[string]ConnectionProfile{"empty": cxProfile}
	cxProfilesTwo := map[string]ConnectionProfile{"empty1": cxProfile, "empty2": cxProfile}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ConnectionProfile
		wantErr bool
	}{
		{name: "test_found", fields: fields{ConnectionProfiles: cxProfiles, Version: "v1.2.3"}, args: args{name: "empty"}, want: &cxProfile, wantErr: false},
		{name: "test_found_one_profile_no_name", fields: fields{ConnectionProfiles: cxProfiles, Version: "v1.2.3"}, args: args{name: ""}, want: &cxProfile, wantErr: false},
		{name: "test_not_found", fields: fields{ConnectionProfiles: cxProfiles, Version: "v1.2.3"}, args: args{name: "other"}, want: nil, wantErr: true},
		{name: "test_no_config", fields: fields{ConnectionProfiles: cxProfiles, Version: "v1.2.3"}, args: args{name: "other"}, want: nil, wantErr: true},
		{name: "test_no_profiles", fields: fields{ConnectionProfiles: map[string]ConnectionProfile{}, Version: "v1.2.3"}, args: args{name: "other"}, want: nil, wantErr: true},
		{name: "test_two_profiles_no_name", fields: fields{ConnectionProfiles: cxProfilesTwo, Version: "v1.2.3"}, args: args{name: ""}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ConnectionProfiles: tt.fields.ConnectionProfiles,
				Version:            tt.fields.Version,
			}
			if tt.name == "test_no_config" {
				c = nil
			}

			got, err := c.GetConnectionProfile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.GetConnectionProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.GetConnectionProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_NewClient(t *testing.T) {
	type fields struct {
		ConnectionProfiles map[string]ConnectionProfile
		Version            string
	}
	cxProfile := ConnectionProfile{}
	cxProfiles := map[string]ConnectionProfile{"empty": cxProfile}
	// cxProfilesTwo := map[string]ConnectionProfile{"empty1": cxProfile, "empty2": cxProfile}
	restClient, err := restclient.NewClient(context.Background(), restclient.ConnectionProfile{}, "config_test/v1.2.3", 600)
	if err != nil {
		panic(err)
	}
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	tests := []struct {
		name          string
		fields        fields
		cxProfileName string
		resName       string
		want          *restclient.RestClient
		wantErr       bool
	}{
		{name: "test_found", fields: fields{ConnectionProfiles: cxProfiles, Version: "v1.2.3"}, cxProfileName: "empty", resName: "config_test", want: restClient, wantErr: false},
		{name: "test_not_found", fields: fields{ConnectionProfiles: cxProfiles, Version: "v1.2.3"}, cxProfileName: "other", want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ConnectionProfiles: tt.fields.ConnectionProfiles,
				Version:            tt.fields.Version,
			}
			got, err := c.NewClient(errorHandler, tt.cxProfileName, tt.resName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil || tt.want == nil {
				if got != tt.want {
					t.Errorf("Config.NewClient() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if ok, diffs := tt.want.Equals(got); !ok {
				t.Errorf(diffs)
			}
		})
	}
}
