data "netapp-ontap_autosupport" "autosupport" {
  # required to know which system to interface with
  cx_profile_name = "cluster6"
}
