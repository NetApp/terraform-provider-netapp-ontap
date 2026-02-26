data "netapp-ontap_fpolicy_external_engine" "protocols_fpolicy_external_engine" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "test"
  svm_name = "tf_acc_svm"
}
