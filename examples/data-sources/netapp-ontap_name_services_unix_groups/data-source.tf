# retrieving all UNIX groups for a SVM
data "netapp-ontap_unix_groups" "all_unix_groups" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
  }
}

# retrieving UNIX groups for a SVM matching the name filter
data "netapp-ontap_unix_groups" "matched_unix_groups" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
    name = "unix*"
  }
}
