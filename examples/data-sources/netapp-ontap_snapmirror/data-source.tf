data "netapp-ontap_snapmirror" "snapmirror" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  destination = {
    path = "snapmirror_dest_svm:snap_dest"
  }
}
