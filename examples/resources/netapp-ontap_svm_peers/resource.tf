resource "netapp-ontap_svm_peers" "svm_peers" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  # name = "svm_peers_test"
  svm = {
    name = "acc_test_peer2"
  }
  peer = {
    svm = {
      name = "acc_test2"
    }
    cluster = {
      name = "swenjuncluster-1"
    }
    peer_cx_profile_name = "cluster3"
  }
  applications = ["snapmirror", "flexcache"]
  # state = "peered"
}
