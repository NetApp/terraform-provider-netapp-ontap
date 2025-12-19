resource "netapp-ontap_fpolicy_event" "protocols_fpolicy_event_cifs" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name            = "event_test"
  svm_name        = "tf_acc_svm"
  protocol        = "cifs"
  file_operations = ["create", "write", "close", "rename", "delete"]
  filters         = ["monitor_ads", "close_with_modification"]
  volume_monitoring = false
}

resource "netapp-ontap_fpolicy_event" "protocols_fpolicy_event_nfs" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name            = "event_test2"
  svm_name        = "tf_acc_svm"
  protocol        = "nfsv4"
  file_operations = ["create", "write", "read", "open", "close"]
  filters         = ["first_read", "first_write", "write_with_size_change"]
  volume_monitoring = true
}
