# creating a nvme namespace
resource "netapp-ontap_nvme_namespace" "example1" {
  cx_profile_name = "cluster4"
  name = "/vol/ns/ns_tf"
  svm_name = "tf_acc_svm"
  ostype = "linux"
  space = {
    size = 10
    size_unit = "kb"
    block_size = 512
  }
}

