resource "netapp-ontap_snapmirror" "snapmirror_async" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  source_endpoint = {
    path = "snapmirror_source_svm:snap"
  }
  destination_endpoint = {
    path = "snapmirror_dest_svm:snap_dest"
  }
  initialize = true
  state = "snapmirrored"
  force = false   # only set to true when breaking the relationship (state = "broken_off")
  transferring_time_out = 600
  quick_resync = false # Only set for resync snapmirror
}