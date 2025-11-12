# retrieving S3 policies for a SVM matching the name filter

data "netapp-ontap_s3_policies" "example1" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    svm_name = "tf_acc_svm"
    name = "test_policy*"
  }
}

# retrieving S3 policies for a SVM
data "netapp-ontap_s3_policies" "example2" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  filter = {
    svm_name = "tf_acc_svm"
  }
}
