data "netapp-ontap_protocols_fpolicy_external_engines" "protocols_fpolicy_external_engines" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  
  filter = {
    svm_name = "svm1",
    name  = "test_engine"
  }
}
