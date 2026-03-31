# retrieving all S3 groups for a SVM
data "netapp-ontap_s3_users" "example1" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
  }
}

# retrieving S3 groups for a SVM matching the name filter
data "netapp-ontap_s3_users" "example2" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  filter = {
    svm_name = "svm1"
    name = "csahu_test*"
  }
}
