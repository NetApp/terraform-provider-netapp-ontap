resource "netapp-ontap_cluster_schedule_resource" "cs_example1" {
  cx_profile_name = "cluster2"
  name = "cs_test_cron"
  cron = {
    minutes = [2, 3]
    hours = [10]
    days = [1]
    months = [6, 7]
    weekdays = [1, 3]
  }
}

resource "netapp-ontap_cluster_schedule_resource" "cs_example2" {
  cx_profile_name = "cluster2"
  name = "cs_test_interval"
  interval = "PT7M30S"
}