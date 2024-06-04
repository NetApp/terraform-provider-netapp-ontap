data "netapp-ontap_svm_peers" "svm_peers" {
  cx_profile_name = "cluster4"
  filter = {
    svm = {
      name = "acc*"
    }
    peer = {
      svm = {
        name = "acc*"
      }
      cluster = {
        name = "abc-1"
      }
    }
  }
}

