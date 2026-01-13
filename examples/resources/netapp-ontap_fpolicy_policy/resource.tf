resource "netapp-ontap_fpolicy_policy" "example_basic" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  svm_name        = "scooby"
  name            = "compliance-policy"
  
  events = ["scooby2"]
  
  scope = {
    # include_shares = ["share1", "share2"]
    # exclude_shares = ["admin_share"]
    include_volumes   = ["scooby_root"]
    check_extensions_on_directories     = true
    object_monitoring_with_no_extension = true
  }
    mandatory        = false
    priority = 2
}

# resource "netapp-ontap_fpolicy_policy" "example_with_scope" {
#   cx_profile_name = "cluster1"
#   svm_name        = "tf_acc_svm"
#   name            = "scoped-compliance-policy"
  
#   events = ["tf_event"]
  
#   mandatory               = false
#   passthrough_read        = false
#   priority                = 5
  
#   # Apply policy to specific volumes and file extensions
#   scope = {
#     include_volumes   = ["vol1", "vol2"]
#     exclude_volumes   = ["temp_vol"]
#     include_extension = ["txt", "doc", "pdf"]
#     exclude_extension = ["tmp", "log"]
#     check_extensions_on_directories     = false
#     object_monitoring_with_no_extension = false
#   }
# }

# resource "netapp-ontap_fpolicy_policy" "example_with_shares" {
#   cx_profile_name = "cluster1"
#   svm_name        = "tf_acc_svm"
#   name            = "share-monitoring-policy"
  
#   engine = "cifs_engine"
#   events = ["tf_event"]
  
#   mandatory        = true
#   passthrough_read = false
  
#   # Apply policy to specific shares
#   scope = {
#     include_shares = ["share1", "share2"]
#     exclude_shares = ["admin_share"]
#   }
# }

# resource "netapp-ontap_fpolicy_policy" "example_with_export_policies" {
#   cx_profile_name = "cluster1"
#   svm_name        = "tf_acc_svm"
#   name            = "nfs-monitoring-policy"
  
#   engine = "nfs_compliance_engine"
#   events = ["tf_event"]
  
#   mandatory        = false
#   passthrough_read = true
  
#   # Apply policy to specific export policies
#   scope = {
#     include_export_policies = ["export_policy1", "export_policy2"]
#     exclude_export_policies = ["test_export"]
#   }
# }
