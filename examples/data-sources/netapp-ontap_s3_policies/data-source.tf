# Example Terraform configuration for multiple S3 Policies data source

data "netapp-ontap_s3_policies" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    svm_name = "tf_acc_svm"
  }
}
