data "netapp-ontap_svm_peer" "svm_peer" {
  cx_profile_name = "cluster4"
  svm = {
    name = "acc_test_peer"
  }
  peer = {
    svm = {
      name = "acc_test"
    }
  }
}

