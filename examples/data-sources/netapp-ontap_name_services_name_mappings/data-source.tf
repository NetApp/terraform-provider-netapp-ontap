# retrieving all name mappings for a SVM
data "netapp-ontap_name_services_name_mappings" "all_name_mappings" {
  # required to know which system to interface with
  cx_profile_name = "cluster"
  filter = {
    svm_name = "svm1"
  }
}

# retrieving name mappings for a SVM and direction
data "netapp-ontap_name_services_name_mappings" "filtered_name_mappings" {
  # required to know which system to interface with
  cx_profile_name = "cluster"
  filter = {
    svm_name = "svm1"
    direction = "unix_win"
  }
}
