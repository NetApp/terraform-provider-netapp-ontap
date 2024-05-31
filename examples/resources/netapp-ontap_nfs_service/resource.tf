resource "netapp-ontap_nfs_service" "protocols_nfs_service" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  svm_name = "ansibleSVM"
  enabled = true
  protocol = {
    v3_enabled = false
    v40_enabled = true
    v40_features = {
      acl_enabled = true
    }
  }
}
