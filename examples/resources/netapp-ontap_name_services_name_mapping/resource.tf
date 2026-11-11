resource "netapp-ontap_name_services_name_mapping" "name_services_name_mapping" {
  # required to know which system to interface with
  cx_profile_name = "cluster"
  svm_name = "svm1"
  index = 4
  direction = "unix_win"
  pattern = "test_pattern"
  replacement = "user"
  client_match = "10.254.101.112/28"

}
