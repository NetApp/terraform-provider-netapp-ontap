data "netapp-ontap_s3_bucket" "s3_bucket" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name = "test-s3-bucket"
  svm_name = "svm1"
}
