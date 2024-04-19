resource "netapp-ontap_cluster_resource" "cluster" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "test_cluster"
  password = "Netapp1!"
  contact = "example@company.com"
  dns_domains = ["domian.netapp.com"]
  name_servers = ["0.0.0.0"]
  timezone = {
    name ="US/Eastern"
  }
  location = "office"
}
