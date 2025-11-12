package interfaces

import (
    "fmt"
    "strings"

    "github.com/mitchellh/mapstructure"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ConvertSIDToString converts SID from interface{} to string
func ConvertSIDToString(sid interface{}) string {
    if sid == nil {
        return ""
    }
    return fmt.Sprintf("%v", sid)
}

// ProtocolsS3PolicyGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsS3PolicyGetDataModelONTAP struct {
    Name       string                              `mapstructure:"name"`
    Comment    string                              `mapstructure:"comment,omitempty"`
    ReadOnly   bool                                `mapstructure:"read_only,omitempty"`
    SVM        svm                                 `mapstructure:"svm"`
    Statements []ProtocolsS3PolicyStatementONTAP   `mapstructure:"statements"`
}

// ProtocolsS3PolicyStatementONTAP describes policy statement structure
type ProtocolsS3PolicyStatementONTAP struct {
    Effect      string                     `mapstructure:"effect"`
    Actions     []string                   `mapstructure:"actions,omitempty"`
    Resources   []string                   `mapstructure:"resources,omitempty"`
    SID         interface{}                `mapstructure:"sid,omitempty"`
}

// ProtocolsS3PolicyResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsS3PolicyResourceBodyDataModelONTAP struct {
    Name       string                              `mapstructure:"name"`
    Comment    string                              `mapstructure:"comment,omitempty"`
    Statements []ProtocolsS3PolicyStatementONTAP   `mapstructure:"statements,omitempty"`
}

// ProtocolsS3PolicyDataSourceFilterModel describes the data source data model for queries.
type ProtocolsS3PolicyFilterModel struct {
    Name    string `mapstructure:"name"`
    SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsS3Policy to get protocols_s3_policy info
func GetProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmUUID string, version versionModelONTAP) (*ProtocolsS3PolicyGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 policy info", "protocols/s3/services/{svm.uuid}/policies/{name} API supported on ONTAP version 9.8 or later")
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmUUID)
    query := r.NewQuery()
    query.Set("name", name)
    query.Fields([]string{"name", "comment", "read_only", "statements", "svm"})

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
func GetProtocolsS3Policies(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, version versionModelONTAP, policyName string) ([]ProtocolsS3PolicyGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 policies info", "protocols/s3/services/{svm.uuid}/policies API supported on ONTAP version 9.8 or later")
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmUUID)
    query := r.NewQuery()
	if policyName != "" {
		query.Set("name", policyName)
	}
    query.Fields([]string{"name", "comment", "read_only", "statements", "svm"})

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
func CreateProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, body ProtocolsS3PolicyResourceBodyDataModelONTAP) (*ProtocolsS3PolicyGetDataModelONTAP, error) {

    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmUUID)
    
    // Manually create the body map with correct lowercase field names
    bodyMap := map[string]interface{}{
        "name": body.Name,
    }
    
    if body.Comment != "" {
        bodyMap["comment"] = body.Comment
    }
    
    if len(body.Statements) > 0 {
        statements := make([]map[string]interface{}, len(body.Statements))
        for i, stmt := range body.Statements {
            // Ensure effect is lowercase for API consistency
            effect := strings.ToLower(stmt.Effect)
            statements[i] = map[string]interface{}{
                "effect": effect,
                "actions": stmt.Actions,
                "resources": stmt.Resources,
                "sid": stmt.SID,
            }
        }
        bodyMap["statements"] = statements
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
    
    bodyMap, ok := body.(map[string]interface{})
    if !ok {
        return errorHandler.MakeAndReportError("error processing update body", "update body must be a map[string]interface{}")
    }
    
    // Ensure effect is lowercase for API compatibility
    if bodyMap["statements"] != nil {
        if statements, ok := bodyMap["statements"].([]interface{}); ok {
            for i, stmt := range statements {
                if stmtMap, ok := stmt.(map[string]interface{}); ok {
                    // Convert effect to lowercase if it exists
                    if effect, exists := stmtMap["effect"]; exists {
                        stmtMap["effect"] = strings.ToLower(fmt.Sprintf("%v", effect))
                    }
                    statements[i] = stmtMap
                }
            }
            bodyMap["statements"] = statements
        }
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
