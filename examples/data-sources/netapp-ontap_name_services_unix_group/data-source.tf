data "netapp-ontap_unix_group" "unix_group" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name = "unix_group1"
  svm_name = "svm1"
}
