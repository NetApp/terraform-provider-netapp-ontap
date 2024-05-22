data "netapp-ontap_protocols_nfs_export_policy_rules" "rules" {
  cx_profile_name = "cluster4"
  svm_name = "ansibleSVM"
  export_policy_name = "default"
  filter = {
    svm_name = "ansibleSVM"
  }
}
