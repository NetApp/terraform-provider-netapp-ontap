data "netapp-ontap_s3_groups" "s3_groups" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    svm_name = "svm1"
  }
}
