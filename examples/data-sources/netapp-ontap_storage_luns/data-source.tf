data "netapp-ontap_storage_luns_data_source" "storage_luns" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  filter = {
    svm_name = "svm0"
  }
}

# data "netapp-ontap_storage_luns_data_source" "storage_luns_not_found" {
#   # required to know which system to interface with
#   cx_profile_name = "cluster4"
#   filter = {
#     svm_name = "svm1"
#   }
# }
