resource "netapp-ontap_protocols_nfs_export_policy_rule_resource" "example" {
  cx_profile_name = "cluster4"
  vserver = "automation"
  export_policy_id = "12884901891"
  clients_match = ["0.0.0.0/0"]
  protocols = ["any"]
  ro_rule =  ["any"]
  rw_rule =  ["any"]
  superuser = ["any"]
  allow_device_creation = true
  chown_mode = "restricted"
}
