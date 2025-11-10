# Example Terraform configuration for multiple S3 Policies data source

data "netapp-ontap_s3_policies" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster6"
  # Optional filter to search for policies by name
  # Priority: filter.svm_name should override top-level svm_name
  svm_name = "tf_acc_svm"
  filter = {
    # svm_name = "tf*"
    name = "test_*"
  }
}
