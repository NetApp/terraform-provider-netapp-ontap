resource "netapp-ontap_snapmirror_resource" "snapmirror_async" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  source_endpoint = {
    path = "snapmirror_source_svm:snap"
  }
  destination_endpoint = {
    path = "snapmirror_dest_svm:snap_dest"
  }
}