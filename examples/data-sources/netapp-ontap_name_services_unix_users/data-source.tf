# retrieving all UNIX users for a SVM
data "netapp-ontap_unix_users" "all_unix_users" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
  }
}

# retrieving UNIX users for a SVM matching the name filter
data "netapp-ontap_unix_users" "matched_unix_users" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
    name = "test*"
  }
}
