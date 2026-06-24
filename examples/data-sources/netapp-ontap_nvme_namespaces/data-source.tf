data "netapp-ontap_nvme_namespaces" "example1" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "tf_acc_svm"
  }
}
