data "netapp-ontap_snapmirrors" "snapmirrors" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    "destination_path" = "snapmirror_dest_svm*"
  }
}
