data "netapp-ontap_snapmirror_policy_data_source" "snapmirror_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "Asynchronous"
}
