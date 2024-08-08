resource "netapp-ontap_cluster_peers" "cluster_peers" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  # name = "testme"
  remote = {
    ip_addresses = ["10.10.10.10"]
  }
  source_details = {
    ip_addresses = ["10.10.10.11"]
  }
  peer_cx_profile_name = "cluster2"
  # generate_passphrase = true
  passphrase = "12345678"
  peer_applications = ["snapmirror"]
}
