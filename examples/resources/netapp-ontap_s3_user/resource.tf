# creating a S3 user configuration with user keys specified
resource "netapp-ontap_s3_user" "example1" {
  cx_profile_name = "hw-cluster"
  name = "test_s3_user"
  svm_name = "svm1"
  comment = "test user"
  access_key = "<AWS-ACCESS-KEY-ID>"
  secret_key = "<AWS-SECRET-ACCESS-KEY>"
}

# regenerating keys and setting new expiry configuration for a specific S3 user
resource "netapp-ontap_s3_user" "example2" {
  cx_profile_name = "hw-cluster"
  name = "test_s3_user"
  svm_name = "svm1"
  comment = "test s3 user"
  regenerate_keys = true
  key_time_to_live = "PT2M"
}

# deleting keys for a specific S3 user
resource "netapp-ontap_s3_user" "example3" {
  cx_profile_name = "hw-cluster"
  name = "test_s3_user"
  svm_name = "svm1"
  comment = "test s3 user"
  delete_keys = true
}
