# Reading a NVMe namespace
data "netapp-ontap_nvme_namespace" "example1" {
  cx_profile_name = "cluster4"
  name = "/vol/ns/ns_tf"
  svm_name = "tf_acc_svm"
}
