package interfaces

import (
    "encoding/json"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/mitchellh/mapstructure"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsS3PolicyGetDataModelONTAP describes the GET data model using go types for mapping.
type ProtocolsS3PolicyGetDataModelONTAP struct {
    Name       string                              `mapstructure:"name"`
    Comment    string                              `mapstructure:"comment,omitempty"`
    Statements []ProtocolsS3PolicyStatementONTAP   `mapstructure:"statements,omitempty"`
    SVM        ProtocolsS3PolicySVMONTAP           `mapstructure:"svm"`
}

// ProtocolsS3PolicyStatementONTAP describes policy statement structure
type ProtocolsS3PolicyStatementONTAP struct {
    Effect      string                     `mapstructure:"effect" json:"effect"`
    Actions     []string                   `mapstructure:"actions,omitempty" json:"actions,omitempty"`
    Resources   []string                   `mapstructure:"resources,omitempty" json:"resources,omitempty"`
    SID         string                     `mapstructure:"sid,omitempty" json:"sid,omitempty"`
}

// ProtocolsS3PolicySVMONTAP describes the SVM data model
type ProtocolsS3PolicySVMONTAP struct {
    Name string `mapstructure:"name" json:"name"`
    UUID string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
}

// ProtocolsS3PolicyResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsS3PolicyResourceBodyDataModelONTAP struct {
    Name       string                              `mapstructure:"name" json:"name"`
    Comment    string                              `mapstructure:"comment,omitempty" json:"comment,omitempty"`
    Statements []ProtocolsS3PolicyStatementONTAP   `mapstructure:"statements,omitempty" json:"statements,omitempty"`
    SVM        ProtocolsS3PolicySVMONTAP           `mapstructure:"svm" json:"-"`
}

// ProtocolsS3PolicyDataSourceModel describes the data source data model.
type ProtocolsS3PolicyDataSourceModel struct {
    CxProfileName types.String                          `tfsdk:"cx_profile_name"`
    Name          types.String                          `tfsdk:"name"`
    Comment       types.String                          `tfsdk:"comment"`
    SVMName       types.String                          `tfsdk:"svm_name"`
    Statements    []ProtocolsS3PolicyStatementModel     `tfsdk:"statements"`
    SVM           ProtocolsS3PolicySVMModel             `tfsdk:"svm"`
    ID            types.String                          `tfsdk:"id"`
}

// ProtocolsS3PolicyStatementModel describes the statement model
type ProtocolsS3PolicyStatementModel struct {
    Effect      types.String              `tfsdk:"effect"`
    Actions     []types.String            `tfsdk:"actions"`
    Resources   []types.String            `tfsdk:"resources"`
    SID         types.String              `tfsdk:"sid"`
}

// ProtocolsS3PolicySVMModel describes the SVM model
type ProtocolsS3PolicySVMModel struct {
    Name types.String `tfsdk:"name"`
    UUID types.String `tfsdk:"uuid"`
}

type ProtocolsS3PolicyDataSourceFilterModel struct {
    Name    types.String `tfsdk:"name"`
    SVMName types.String `tfsdk:"svm_name"`
}

// GetProtocolsS3Policy to get protocols_s3_policy info
func GetProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*ProtocolsS3PolicyGetDataModelONTAP, error) {
    // First get SVM UUID
    svmAPI := "svm/svms"
    svmQuery := r.NewQuery()
    svmQuery.Set("name", svmName)
    svmQuery.Fields([]string{"uuid"})
    
    statusCode, svmResponse, err := r.GetNilOrOneRecord(svmAPI, svmQuery, nil)
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET %s: %s, statusCode %d", svmAPI, err, statusCode))
    }
    if svmResponse == nil {
        return nil, errorHandler.MakeAndReportError("svm not found", fmt.Sprintf("svm %s not found", svmName))
    }
    
    var svmData map[string]interface{}
    if err := mapstructure.Decode(svmResponse, &svmData); err != nil {
        return nil, errorHandler.MakeAndReportError("error decoding svm info", fmt.Sprintf("error decoding svm info: %s", err))
    }
    
    svmUUID, ok := svmData["uuid"].(string)
    if !ok {
        return nil, errorHandler.MakeAndReportError("error getting svm uuid", "could not extract svm uuid")
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmUUID)
    query := r.NewQuery()
    query.Set("name", name)
    query.Fields([]string{"name", "comment", "statements", "svm.name", "svm.uuid"})

    statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
    if err == nil && response == nil {
        err = fmt.Errorf("protocols_s3_policy %s not found", name)
    }
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error reading protocols_s3_policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
    }

    var dataONTAP ProtocolsS3PolicyGetDataModelONTAP
    if err := mapstructure.Decode(response, &dataONTAP); err != nil {
        return nil, errorHandler.MakeAndReportError("error decoding protocols_s3_policy info", fmt.Sprintf("error on decode protocols_s3_policy info: %s, statusCode %d, response %#v", err, statusCode, response))
    }

    return &dataONTAP, nil
}

// GetProtocolsS3Policies to get protocols_s3_policy info for all resources matching a filter
func GetProtocolsS3Policies(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsS3PolicyDataSourceFilterModel) ([]ProtocolsS3PolicyGetDataModelONTAP, error) {
    // First get SVM UUID
    svmAPI := "svm/svms"
    svmQuery := r.NewQuery()
    svmQuery.Set("name", filter.SVMName.ValueString())
    svmQuery.Fields([]string{"uuid"})
    
    statusCode, svmResponse, err := r.GetNilOrOneRecord(svmAPI, svmQuery, nil)
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET %s: %s, statusCode %d", svmAPI, err, statusCode))
    }
    if svmResponse == nil {
        return nil, errorHandler.MakeAndReportError("svm not found", fmt.Sprintf("svm %s not found", filter.SVMName.ValueString()))
    }
    
    var svmData map[string]interface{}
    if err := mapstructure.Decode(svmResponse, &svmData); err != nil {
        return nil, errorHandler.MakeAndReportError("error decoding svm info", fmt.Sprintf("error decoding svm info: %s", err))
    }
    
    svmUUID, ok := svmData["uuid"].(string)
    if !ok {
        return nil, errorHandler.MakeAndReportError("error getting svm uuid", "could not extract svm uuid")
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmUUID)
    query := r.NewQuery()
    query.Fields([]string{"name", "comment", "statements", "svm.name", "svm.uuid"})

    // Only add name filter if explicitly specified
    if filter != nil && !filter.Name.IsNull() && filter.Name.ValueString() != "" {
        query.Set("name", filter.Name.ValueString())
    }

    statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error reading protocols_s3_policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
    }

    var dataONTAP []ProtocolsS3PolicyGetDataModelONTAP
    for _, info := range response {
        var record ProtocolsS3PolicyGetDataModelONTAP
        if err := mapstructure.Decode(info, &record); err != nil {
            return nil, errorHandler.MakeAndReportError("error decoding protocols_s3_policy info", fmt.Sprintf("error on decode protocols_s3_policy info: %s, statusCode %d, info %#v", err, statusCode, info))
        }
        dataONTAP = append(dataONTAP, record)
    }

    return dataONTAP, nil
}

// CreateProtocolsS3Policy to create protocols_s3_policy
func CreateProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsS3PolicyResourceBodyDataModelONTAP, svmName string) (*ProtocolsS3PolicyGetDataModelONTAP, error) {
    // First get SVM UUID
    svmAPI := "svm/svms"
    svmQuery := r.NewQuery()
    svmQuery.Set("name", svmName)
    svmQuery.Fields([]string{"uuid"})
    
    statusCode, svmResponse, err := r.GetNilOrOneRecord(svmAPI, svmQuery, nil)
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET %s: %s, statusCode %d", svmAPI, err, statusCode))
    }
    if svmResponse == nil {
        return nil, errorHandler.MakeAndReportError("svm not found", fmt.Sprintf("svm %s not found", svmName))
    }
    
    var svmData map[string]interface{}
    if err := mapstructure.Decode(svmResponse, &svmData); err != nil {
        return nil, errorHandler.MakeAndReportError("error decoding svm info", fmt.Sprintf("error decoding svm info: %s", err))
    }
    
    svmUUID, ok := svmData["uuid"].(string)
    if !ok {
        return nil, errorHandler.MakeAndReportError("error getting svm uuid", "could not extract svm uuid")
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmUUID)
    
    // Use JSON marshal/unmarshal to respect JSON tags (including json:"-")
    bodyJSON, err := json.Marshal(body)
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error marshaling protocols_s3_policy body", fmt.Sprintf("error marshaling body: %s, body: %#v", err, body))
    }
    
    var bodyMap map[string]interface{}
    if err := json.Unmarshal(bodyJSON, &bodyMap); err != nil {
        return nil, errorHandler.MakeAndReportError("error unmarshaling protocols_s3_policy body", fmt.Sprintf("error unmarshaling body: %s", err))
    }

    query := r.NewQuery()
    query.Add("return_records", "true")

    statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
    if err != nil {
        return nil, errorHandler.MakeAndReportError("error creating protocols_s3_policy", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
    }

    var dataONTAP ProtocolsS3PolicyGetDataModelONTAP
    if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
        return nil, errorHandler.MakeAndReportError("error decoding protocols_s3_policy info", fmt.Sprintf("error on decode protocols_s3_policy info: %s, statusCode %d, response %#v", err, statusCode, response))
    }

    return &dataONTAP, nil
}

// UpdateProtocolsS3Policy to update protocols_s3_policy
func UpdateProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, policyName string, body interface{}) error {
    api := fmt.Sprintf("protocols/s3/services/%s/policies/%s", svmUUID, policyName)
    
    // Use JSON marshal/unmarshal to respect JSON tags (including json:"-")
    bodyJSON, err := json.Marshal(body)
    if err != nil {
        return errorHandler.MakeAndReportError("error marshaling protocols_s3_policy body", fmt.Sprintf("error marshaling body: %s, body: %#v", err, body))
    }
    
    var bodyMap map[string]interface{}
    if err := json.Unmarshal(bodyJSON, &bodyMap); err != nil {
        return errorHandler.MakeAndReportError("error unmarshaling protocols_s3_policy body", fmt.Sprintf("error unmarshaling body: %s", err))
    }

    statusCode, _, err := r.CallUpdateMethod(api, nil, bodyMap)
    if err != nil {
        return errorHandler.MakeAndReportError("error updating protocols_s3_policy", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
    }

    return nil
}

// DeleteProtocolsS3Policy to delete protocols_s3_policy
func DeleteProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, policyName string) error {
    api := fmt.Sprintf("protocols/s3/services/%s/policies/%s", svmUUID, policyName)

    statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
    if err != nil {
        return errorHandler.MakeAndReportError("error deleting protocols_s3_policy", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
    }

    return nil
}
