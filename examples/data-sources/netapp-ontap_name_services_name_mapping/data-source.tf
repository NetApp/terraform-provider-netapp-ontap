data "netapp-ontap_name_services_name_mapping" "name_mapping" {
  # required to know which system to interface with
  cx_profile_name = "cluster"
  svm_name = "svm1"
  direction = "unix_win"
  index = 4
}
