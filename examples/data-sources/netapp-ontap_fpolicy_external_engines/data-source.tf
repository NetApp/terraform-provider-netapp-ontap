data "netapp-ontap_fpolicy_external_engines" "protocols_fpolicy_external_engines" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  
  filter = {
    svm_name = "tf_acc_svm",
    name  = "test"
  }
}
