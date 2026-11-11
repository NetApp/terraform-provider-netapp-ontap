data "netapp-ontap_quota_rules" "storage_quota_rules" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  filter = {
    type = "tree"
    svm_name = "carchi-test"
  }
}
