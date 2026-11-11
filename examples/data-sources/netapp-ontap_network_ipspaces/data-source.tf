# retrieving all IPspaces
data "netapp-ontap_network_ipspaces" "example1" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
}

# retrieving IPspaces matching the name filter
data "netapp-ontap_network_ipspaces" "example2" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    name = "test*"
  }
}
