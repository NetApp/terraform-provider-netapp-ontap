package interfaces

import (
    "fmt"
    "strings"

    "github.com/hashicorp/terraform-plugin-log/tflog"
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
type ProtocolsS3PolicyDataSourceFilterModel struct {
    Name    string `mapstructure:"name"`
    SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsS3Policy to get protocols_s3_policy info
func GetProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*ProtocolsS3PolicyGetDataModelONTAP, error) {
    // Get SVM info using existing function
    svmInfo, err := GetSvmByName(errorHandler, r, svmName)
    if err != nil {
        return nil, err
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmInfo.UUID)
    query := r.NewQuery()
    query.Set("name", name)
    query.Fields([]string{"name", "comment", "read_only", "statements", "svm.name"})

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

// GetProtocolsS3PolicyByFilter to get protocols_s3_policy info by filter
func GetProtocolsS3PolicyByFilter(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsS3PolicyDataSourceFilterModel) (*ProtocolsS3PolicyGetDataModelONTAP, error) {
    // Get SVM info using existing function
    svmInfo, err := GetSvmByName(errorHandler, r, filter.SVMName)
    if err != nil {
        return nil, err
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmInfo.UUID)
    query := r.NewQuery()
    query.Fields([]string{"name", "comment", "read_only", "statements", "svm.name"})

    if filter != nil {
        var filterMap map[string]interface{}
        if err := mapstructure.Decode(filter, &filterMap); err != nil {
            return nil, errorHandler.MakeAndReportError("error encoding protocols_s3_policy filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
        }
        // Remove svm.name from filter since we use it for SVM lookup
        delete(filterMap, "svm.name")
        query.SetValues(filterMap)
    }

    statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
    if err == nil && response == nil {
        err = fmt.Errorf("protocols_s3_policy not found")
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
    // If SVM name contains wildcards or we want to search across SVMs
    if filter != nil && filter.SVMName != "" && !containsWildcard(filter.SVMName) {
        return GetProtocolsS3PoliciesFromSVM(errorHandler, r, filter)
    }
    
    // For wildcard SVM names or cross-SVM queries, we need to get matching SVMs first
    svmFilter := &SvmDataSourceFilterModel{}
    if filter != nil && filter.SVMName != "" {
        svmFilter.Name = filter.SVMName
    }
    
    // Get matching SVMs
    svms, err := GetSvmsByName(errorHandler, r, svmFilter, versionModelONTAP{Generation: 9, Major: 13, Minor: 0}) // Use compatible version
    if err != nil {
        return nil, err
    }
    
    var allPolicies []ProtocolsS3PolicyGetDataModelONTAP
    
    // Query each matching SVM for S3 policies
    for _, svm := range svms {
        svmFilter := &ProtocolsS3PolicyDataSourceFilterModel{
            SVMName: svm.Name,
            Name:    "",
        }
        if filter != nil {
            svmFilter.Name = filter.Name
        }
        
        policies, err := GetProtocolsS3PoliciesFromSVM(errorHandler, r, svmFilter)
        if err != nil {
            // Log error but continue with other SVMs - SVM might not have S3 service enabled
            tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Could not get S3 policies from SVM %s: %s", svm.Name, err.Error()))
            continue
        }
        
        allPolicies = append(allPolicies, policies...)
    }
    
    return allPolicies, nil
}

// containsWildcard checks if a string contains wildcard characters
func containsWildcard(s string) bool {
    return strings.Contains(s, "*") || strings.Contains(s, "?")
}

// GetProtocolsS3PoliciesFromSVM to get protocols_s3_policy info from a specific SVM
func GetProtocolsS3PoliciesFromSVM(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsS3PolicyDataSourceFilterModel) ([]ProtocolsS3PolicyGetDataModelONTAP, error) {
    // Get SVM info using existing function
    svmInfo, err := GetSvmByName(errorHandler, r, filter.SVMName)
    if err != nil {
        return nil, err
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmInfo.UUID)
    query := r.NewQuery()
    query.Fields([]string{"name", "comment", "read_only", "statements", "svm.name"})

    if filter != nil {
        var filterMap map[string]interface{}
        if err := mapstructure.Decode(filter, &filterMap); err != nil {
            return nil, errorHandler.MakeAndReportError("error encoding protocols_s3_policy filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
        }
        delete(filterMap, "svm.name")
        query.SetValues(filterMap)
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
    // Get SVM info using existing function
    svmInfo, err := GetSvmByName(errorHandler, r, svmName)
    if err != nil {
        return nil, err
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies", svmInfo.UUID)
    
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
func UpdateProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, policyName string, body interface{}) error {
    // Get SVM info using existing function
    svmInfo, err := GetSvmByName(errorHandler, r, svmName)
    if err != nil {
        return err
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies/%s", svmInfo.UUID, policyName)
    
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
func DeleteProtocolsS3Policy(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, policyName string) error {
    // Get SVM info using existing function
    svmInfo, err := GetSvmByName(errorHandler, r, svmName)
    if err != nil {
        return err
    }
    
    api := fmt.Sprintf("protocols/s3/services/%s/policies/%s", svmInfo.UUID, policyName)

    statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
    if err != nil {
        return errorHandler.MakeAndReportError("error deleting protocols_s3_policy", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
    }

    return nil
}
