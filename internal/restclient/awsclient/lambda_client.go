package awsclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AWSLambdaClient struct {
	Lambda  *lambda.Client
	ctx     context.Context
	profile AWSLambdaProfile
}

type AWSLambdaProfile struct {
	APIRoot  string
	Username string
	Password string
	Hostname string
	// ValidateCerts bool
	Base64Credential string
	AWSConfig        AWSConfig `mapstructure:"aws,omitempty"`
}

type AWSConfig struct {
	FunctionName        string
	Region              string
	SharedConfigProfile string
}

type RequestType string

const (
	HTTPS  RequestType = "https"
	HEALTH RequestType = "health"
)

type Payload struct {
	Body Body `json:"body"`
}

type Body struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Data        interface{}       `json:"data,omitempty"`
	QueryParams interface{}       `json:"queryParams,omitempty"`
	Endpoint    string            `json:"endpoint"`
	Headers     map[string]string `json:"headers"`
	RequestType RequestType       `json:"requestType"`
}

// NewClient creates a new AWS Lambda client
func NewClient(ctx context.Context, profile AWSLambdaProfile) (*AWSLambdaClient, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(profile.AWSConfig.Region), config.WithSharedConfigProfile(profile.AWSConfig.SharedConfigProfile))
	if err != nil {
		return nil, err
	}
	lambdaClient := lambda.NewFromConfig(sdkConfig)

	encodedCredential := createBase64Credential(profile.Username, profile.Password)
	profile.Base64Credential = encodedCredential

	return &AWSLambdaClient{
		Lambda:  lambdaClient,
		profile: profile,
		ctx:     ctx,
	}, nil
}

// Invoke sends the API Request to the AWS Lambda function
func (c *AWSLambdaClient) Invoke(baseURL string, method string, body map[string]interface{}, queryValues url.Values) (int, []byte, error) {
	statusCode := -1
	query := make(map[string]string)

	if len(queryValues) > 0 {
		for key, value := range queryValues {
			query[key] = value[0]
		}
	}
	payloadStruct := constructPayload("api/"+baseURL, method, c.profile.Hostname, body, query, c.profile.Base64Credential)
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("Error marshalling payload:%#v", err))
		return statusCode, nil, err
	}
	invokeOutput, err := c.Lambda.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(c.profile.AWSConfig.FunctionName),
		LogType:      types.LogTypeTail,
		Payload:      payloadBytes,
	})
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("Error invoking Lambda function:%#v", err))
		return statusCode, nil, err
	}
	return int(invokeOutput.StatusCode), invokeOutput.Payload, nil
}

func constructPayload(url, method, endpoint string, body, queryParams interface{}, authorizationHeader string) Payload {
	// Add the Authorization header if present
	headers := make(map[string]string)
	if authorizationHeader != "" {
		headers["Authorization"] = "Basic " + authorizationHeader
	}
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "*/*"
	payload := Payload{
		Body: Body{
			URL:         url,
			Method:      method,
			Data:        body,
			QueryParams: queryParams,
			Endpoint:    endpoint,
			Headers:     headers,
			RequestType: HTTPS, // Assuming HTTPS request type for this example
		},
	}

	return payload
}

func createBase64Credential(username, password string) string {
	// Concatenate username and password with a colon
	credential := fmt.Sprintf("%s:%s", username, password)

	// Encode the credential string to Base64
	base64Credential := base64.StdEncoding.EncodeToString([]byte(credential))

	return base64Credential
}
