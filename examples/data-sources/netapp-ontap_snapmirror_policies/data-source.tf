data "netapp-ontap_snapmirror_policies_data_source" "snapmirror_policies" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "Async*"
  }
}
