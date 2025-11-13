data "netapp-ontap_s3_group" "s3_group" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "csahu_test_group"
  svm_name = "svm1"
}
