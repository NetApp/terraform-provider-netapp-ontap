resource "netapp-ontap_s3_bucket" "s3_bucket" {
  cx_profile_name = "hw-cluster"
  name = "test-s3-bucket"
  svm_name = "svm1"
  size = 199229440
  comment = "test s3 bucket"
  qos_policy = {
    max_throughput_iops = 100
    max_throughput_mbps = 150
  }
  policy = {
    statements = [
      {
        sid = "AccessToUser1"
        resources = [
          "bucket1/",
          "bucket1/*"
        ]
        actions = [
          "GetObject"
        ]
        effect = "allow"
        conditions = [
          {
            operator = "string_like"
            usernames = [
              "user"
            ]
          }
        ]
        principals = [
          "user1"
        ]
      }
    ]
  }
}
