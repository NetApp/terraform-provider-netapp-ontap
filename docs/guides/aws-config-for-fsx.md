---
page_title: "Setup to use Amazon FSx for Netapp ONTAP via AWS Lambda"
subcategory: ""
description: |-
---

# Setup to use Amazon FSx for Netapp ONTAP via AWS Lambda

This page is specifically for FSx ONTAP system, please skip this part if you are not the target audience. In this provider, AWS Lambda is used as gateway from the client side to communicate with FSx ONTAP, which resides in AWS EC2 instances.

# Create a Lambda Link
[Create a link](https://docs.netapp.com/us-en/workload-fsx-ontap/create-link.html) and use the link name in the connection profile.

# Shared config and credentials files
The shared AWS config and credentials files contain a set of profiles. A profile is a set of configuration settings, in key–value pairs, that is used by the AWS Command Line Interface (AWS CLI), the AWS SDKs, and other tools. Configuration values are attached to a profile in order to configure some aspect of the SDK/tool when that profile is used. These files are "shared" in that the values take affect for any applications, processes, or SDKs on the local environment for a user.
[AWS shared config and credentials file format](https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html)
[AWS shared config and credentials file location](https://docs.aws.amazon.com/sdkref/latest/guide/file-location.html)

You must setup credentials files to use Fsx ONTAP.
Example of AWS credentials file
[fsx]
aws_access_key_id = <aws_access_key_id>
aws_secret_access_key = <aws_secret_access_key>

You can either specify `region` in the `connection_profiles` or in the AWS config files. If both palces have region set up, the region in the Terraform config file will be used. The profile name must be the same in both AWS config and credentials files.
Example of region dlecaration in AWS config profile
[profile fsx]
region = us-east-1

# Construct connection profiles
```
provider "netapp-ontap" {
  # A connection profile defines how to interface with an ONTAP cluster or svm.
  # At least one is required.
  connection_profiles = [
    {
      name = "fsx"
      hostname = "aws.management.endpoint.com" #the management endpoints for the FSxN system.
      username = "admin"
      password = "Password"
      aws = {
        function_name = "lambda_link_name"
        region = "aws_region"
        shared_config_profile = "fsx"
      }
    }
  ]
}
```