# retrieving all S3 buckets for a SVM
data "netapp-ontap_s3_buckets" "s3_buckets_example1" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
  }
}

# retrieving S3 buckets for a SVM matching the name filter
data "netapp-ontap_s3_buckets" "s3_buckets_example2" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
    name = "test*"
  }
}
