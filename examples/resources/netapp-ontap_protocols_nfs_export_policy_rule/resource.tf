resource "netapp-ontap_protocols_nfs_export_policy_rule" "example" {
  cx_profile_name = "cluster4"
  svm_name = "svm0"
  export_policy_name = "export_policy"
  clients_match = ["0.0.0.0/0"]
  protocols = ["any"]
  ro_rule =  ["any"]
  rw_rule =  ["none"]
  superuser = ["none"]
}
