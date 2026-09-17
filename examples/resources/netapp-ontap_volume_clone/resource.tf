# cloning a volume
resource "netapp-ontap_volume_clone" "volume_clone" {
  cx_profile_name = "hw-cluster"
  name = "vol1_clone1"
  svm_name = "svm1"
  type = "rw"
  clone = {
    parent_volume = "vol1"
  }
  nas = {
    junction_path = "/cloned_volume"
    group_id = 1
    user_id = 1
  }
}

# splitting a volume clone
resource "netapp-ontap_volume_clone" "volume_clone" {
  cx_profile_name = "hw-cluster"
  name = "vol1_clone1"
  svm_name = "svm1"
  type = "rw"
  clone = {
    parent_volume = "vol1"
    split = true
  }
}

# cloning and splitting a volume using a snapshot
resource "netapp-ontap_volume_clone" "volume_clone" {
  cx_profile_name = "hw-cluster"
  name = "vol1_clone1"
  svm_name = "svm1"
  type = "rw"
  clone = {
    parent_volume = "vol1"
    parent_snapshot = "daily.2026-07-27_0010"
    split = true
  }
  nas = {
    junction_path = "/cloned_volume"
    group_id = 1
    user_id = 1
  }
}
