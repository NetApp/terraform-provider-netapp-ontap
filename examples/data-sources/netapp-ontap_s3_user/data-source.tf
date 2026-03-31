data "netapp-ontap_s3_user" "s3_user" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name = "test_s3_user"
  svm_name = "svm1"
}
