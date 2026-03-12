package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsS3BucketGetDataModelONTAP describes the GET record data model using go types for mapping
type ProtocolsS3BucketGetDataModelONTAP struct {
	SVM                  svm                          `mapstructure:"svm"`
	Name                 string                       `mapstructure:"name"`
	Size                 int64                        `mapstructure:"size"`
	Comment              string                       `mapstructure:"comment"`
	Type                 string                       `mapstructure:"type"`
	NASPath              string                       `mapstructure:"nas_path"`
	VersioningState      string                       `mapstructure:"versioning_state"`
	Policy              *PolicyDataModel              `mapstructure:"policy,omitempty"`
	QoSPolicy           *QoSPolicyDataModel           `mapstructure:"qos_policy,omitempty"`
	SnapshotPolicy       SnapshotPolicy               `mapstructure:"snapshot_policy"`
	AuditEventSelector  *AuditEventSelectorDataModel  `mapstructure:"audit_event_selector,omitempty"`
	UUID                 string                       `mapstructure:"uuid"`
}

// SnapshotPolicy struct defined in interfaces/svm.go

// PolicyDataModel describes the policy data model using go types for mapping
type PolicyDataModel struct {
	Statements  []StatementsDataModel  `mapstructure:"statements"`
}

// StatementsDataModel describes the statements data model using go types for mapping
type StatementsDataModel struct {
	SID           string               `mapstructure:"sid,omitempty"        json:"sid,omitempty"`
	Resources   []string               `mapstructure:"resources,omitempty"  json:"resources,omitempty"`
	Actions     []string               `mapstructure:"actions"              json:"actions"`
	Effect        string               `mapstructure:"effect,omitempty"     json:"effect,omitempty"`
	Principals  []string               `mapstructure:"principals"           json:"principals"`
	Conditions  []ConditionsDataModel  `mapstructure:"conditions"           json:"conditions"`
}

// ConditionsDataModel describes the conditions data model using go types for mapping
type ConditionsDataModel struct {
	Operator      string  `mapstructure:"operator"              json:"operator"`
	MaxKeys     []string  `mapstructure:"max_keys,omitempty"    json:"max_keys,omitempty"`
	Delimiters  []string  `mapstructure:"delimiters,omitempty"  json:"delimiters,omitempty"`
	SourceIPs   []string  `mapstructure:"source_ips,omitempty"  json:"source_ips,omitempty"`
	Prefixes    []string  `mapstructure:"prefixes,omitempty"    json:"prefixes,omitempty"`
	Usernames   []string  `mapstructure:"usernames,omitempty"   json:"usernames,omitempty"`
}

// QoSPolicyDataModel describes the qos_policy data model using go types for mapping
type QoSPolicyDataModel struct {
	Name               string  `mapstructure:"name,omitempty"`
	MaxThroughputIops  int     `mapstructure:"max_throughput_iops,omitempty"`
	MaxThroughputMbps  int     `mapstructure:"max_throughput_mbps,omitempty"`
	MinThroughputIops  int     `mapstructure:"min_throughput_iops,omitempty"`
	MinThroughputMbps  int     `mapstructure:"min_throughput_mbps,omitempty"`
}

// AuditEventSelectorDataModel describes the audit_event_selector data model using go types for mapping
type AuditEventSelectorDataModel struct {
	Access      string  `mapstructure:"access,omitempty"`
	Permission  string  `mapstructure:"permission,omitempty"`
}

// ProtocolsS3BucketResourceBodyDataModel describes the resource body data model using go types for mapping
type ProtocolsS3BucketResourceBodyDataModel struct {
	Name                        string                       `mapstructure:"name"`
	SVM                         svm                          `mapstructure:"svm"`
	Size                        int64                        `mapstructure:"size,omitempty"`
	Comment                    *string                       `mapstructure:"comment,omitempty"`
	Type                        string                       `mapstructure:"type,omitempty"`
	NASPath                     string                       `mapstructure:"nas_path,omitempty"`
	VersioningState             string                       `mapstructure:"versioning_state,omitempty"`
	Policy                     *PolicyDataModel              `mapstructure:"policy,omitempty"`
	QoSPolicy                  *QoSPolicyDataModel           `mapstructure:"qos_policy,omitempty"`
	SnapshotPolicy              SnapshotPolicy               `mapstructure:"snapshot_policy,omitempty"`
	AuditEventSelector         *AuditEventSelectorDataModel  `mapstructure:"audit_event_selector,omitempty"`
	Aggregates                []string                       `mapstructure:"aggregates,omitempty"`
	ConstituentsPerAggregate    int64                        `mapstructure:"constituents_per_aggregate,omitempty"`
}

// S3BucketsFilterModel describes the filter model for S3 buckets
type S3BucketsFilterModel struct {
	SVMName  string  `mapstructure:"svm.name"`
	Name     string  `mapstructure:"name"`
}

// GetProtocolsS3Bucket to get S3 bucket info given its name
func GetProtocolsS3Bucket(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string, version versionModelONTAP) (*ProtocolsS3BucketGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 bucket info", "protocols/s3/buckets API supported on ONTAP version 9.8 or later")
	}
	api := "protocols/s3/buckets"

	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)
	var fields = []string{
		"name",
		"svm",
		"size",
		"comment",
		"policy",
		"policy.statements",
		"qos_policy",
	}
	if version.Generation == 9 && version.Major >= 10 {
		fields = append(fields, "audit_event_selector")
	}
	if version.Generation == 9 && version.Major >= 11 {
		fields = append(fields, "versioning_state")
	}
	if version.Generation == 9 && version.Major >= 12 {
		fields = append(fields, "type", "nas_path")
	}
	if version.Generation == 9 && version.Major >= 16 {
		fields = append(fields, "snapshot_policy")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading S3 bucket info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsS3BucketGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_bucket data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsS3Buckets to get a list of S3 buckets info for given filter options
func GetProtocolsS3Buckets(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *S3BucketsFilterModel, version versionModelONTAP) ([]ProtocolsS3BucketGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 buckets info", "protocols/s3/buckets API supported on ONTAP version 9.8 or later")
	}
	api := "protocols/s3/buckets"

	query := r.NewQuery()
	var fields = []string{
		"name",
		"svm",
		"size",
		"comment",
		"policy",
		"policy.statements",
		"qos_policy",
	}
	if version.Generation == 9 && version.Major >= 10 {
		fields = append(fields, "audit_event_selector")
	}
	if version.Generation == 9 && version.Major >= 11 {
		fields = append(fields, "versioning_state")
	}
	if version.Generation == 9 && version.Major >= 12 {
		fields = append(fields, "type", "nas_path")
	}
	if version.Generation == 9 && version.Major >= 16 {
		fields = append(fields, "snapshot_policy")
	}
	query.Fields(fields)

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding S3 buckets filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading S3 buckets info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsS3BucketGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsS3BucketGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_buckets data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsS3Bucket to create a S3 bucket
func CreateProtocolsS3Bucket(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3BucketResourceBodyDataModel) (*ProtocolsS3BucketGetDataModelONTAP, error) {
	api := "protocols/s3/buckets"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding protocols/s3/buckets body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating S3 bucket", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsS3BucketGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols/s3/buckets info", fmt.Sprintf("error on decode protocols/s3/buckets info: %s, statusCode %d, response %#v", err, statusCode, response))
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created s3_bucket source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteProtocolsS3Bucket to delete a S3 bucket
func DeleteProtocolsS3Bucket(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, svmUUID string) error {
	api := fmt.Sprintf("protocols/s3/buckets/%s/%s", svmUUID, uuid)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting S3 bucket", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateProtocolsS3Bucket to modify a S3 bucket
func UpdateProtocolsS3Bucket(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3BucketResourceBodyDataModel, uuid string, svmUUID string) error {
	api := fmt.Sprintf("protocols/s3/buckets/%s/%s", svmUUID, uuid)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding protocols/s3/buckets/{svm.uuid}/{uuid} body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error modifying S3 bucket", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
