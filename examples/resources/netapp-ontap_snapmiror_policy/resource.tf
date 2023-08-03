resource "netapp-ontap_snapmirror_policy_resource" "snapmirror_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testsp_basic"
  svm_name = "ansibleSVM"
}

resource "netapp-ontap_snapmirror_policy_resource" "snapmirror_policy_async" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testsp_async_retention"
  svm_name = "ansibleSVM"
  type = "async"
  retention = [{
    label = "weekly"
    count = 2
  },
  {
    label = "daily",
    count = 7
  },
  {
    label = "newlabel1",
    count = 3
  }
  ]
}

resource "netapp-ontap_snapmirror_policy_resource" "snapmirror_policy_sync" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testsp_sync"
  svm_name = "ansibleSVM"
  type = "sync"
  sync_type = "sync"
}

resource "netapp-ontap_snapmirror_policy_resource" "snapmirror_policy_sync_1" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testsp_sync"
  svm_name = "ansibleSVM"
  type = "sync"
  sync_type = "sync"
  retention = [{
    count = 1
    label = "hourly"
  }]
}