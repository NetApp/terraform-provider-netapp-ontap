data "netapp-ontap_svm_peers_data_source" "svm_peers" {
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

