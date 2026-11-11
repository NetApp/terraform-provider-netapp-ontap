data "netapp-ontap_unix_user" "unix_user" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name = "root"
  svm_name = "svm1"
}
