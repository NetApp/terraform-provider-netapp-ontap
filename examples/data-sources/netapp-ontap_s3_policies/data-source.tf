# Example Terraform configuration for multiple S3 Policies data source

data "netapp-ontap_s3_policies" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster6"
  svm_name = "inter_svm"
  # Optional filter to search for policies by name
  filter = {
    name = "test_policy*"
  }
}
