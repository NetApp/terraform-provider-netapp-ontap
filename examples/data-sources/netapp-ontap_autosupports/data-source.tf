data "netapp-ontap_autosupports" "autosupports" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  # filter = {}
}
