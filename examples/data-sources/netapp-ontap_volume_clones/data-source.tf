# retrieving all volume clones for a SVM
data "netapp-ontap_volume_clones" "all_volume_clones" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
  }
}

# retrieving all volume clones for a SVM matching the name filter
data "netapp-ontap_volume_clones" "matched_volume_clones" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
    name = "clone*"
  }
}