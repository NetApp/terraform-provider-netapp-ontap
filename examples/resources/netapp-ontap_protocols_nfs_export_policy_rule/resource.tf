resource "netapp-ontap_protocols_nfs_export_policy_rule_resource" "example" {
  cx_profile_name = "cluster4"
  svm_name = "automation"
  export_policy_name = "test"
  clients_match = ["0.0.0.0/0"]
  protocols = ["any"]
  ro_rule =  ["any"]
  rw_rule =  ["none"]
  superuser = ["none"]
}
