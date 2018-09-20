package models

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/spf13/viper"
)


func CreateSNSEndpoint(device_platform string, device_token string) string {

	platform_arn := ""
	if device_platform == "Android" {
		platform_arn = viper.GetString("SNS_ANDROID_ARN")
	} else if device_platform == "iOS" {
		platform_arn = viper.GetString("SNS_IOS_ARN")
	}

	end_point_arn := ""

	aws_config := aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      aws.String(viper.GetString("SNS_REGION")),
	}

	sess, err := session.NewSession(&aws_config)
	if err != nil {
		return end_point_arn
	}

	if sess == nil {
		return end_point_arn
	}

	sns_client := sns.New(sess, &aws_config)
	if sns_client == nil {
		return end_point_arn
	}

	end_point, err := sns_client.CreatePlatformEndpoint(&sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: aws.String(platform_arn),
		Token: aws.String(device_token),
	})

	if err != nil {
		return end_point_arn
	}

	end_point_arn = aws.StringValue(end_point.EndpointArn)

	return end_point_arn
}


func CreateSNSEndpointWithLambda(device_platform string, device_token string) string {

	platform_arn := ""
	if device_platform == "Android" {
		platform_arn = viper.GetString("SNS_ANDROID_ARN")
	} else if device_platform == "iOS" {
		platform_arn = viper.GetString("SNS_IOS_ARN")
	}

	end_point_arn := ""

	aws_config := aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      aws.String(viper.GetString("SNS_REGION")),
	}

	sess, err := session.NewSession(&aws_config)
	if err != nil {
		return end_point_arn
	}

	if sess == nil {
		return end_point_arn
	}

	lambda_client := lambda.New(sess)
	if lambda_client == nil {
		return end_point_arn
	}

	payload_string := fmt.Sprintf(`{"PlatformApplicationArn":"%s","Token":"%s"}`, platform_arn, device_token)
	payload := []byte(payload_string)

	resp, err := lambda_client.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String("snsSubscribe"),
		Payload: payload,
	})

	if err != nil {
		return end_point_arn
	}

	r := struct {
		EndpointArn string `json:"EndpointArn"`
	} {
		"",
	}

	json.Unmarshal(resp.Payload, &r)

	end_point_arn = r.EndpointArn

	return end_point_arn
}