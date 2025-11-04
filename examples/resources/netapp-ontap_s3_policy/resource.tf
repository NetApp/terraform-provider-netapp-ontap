# Example without statements
resource "netapp-ontap_s3_policy" "basic_policy" {
  cx_profile_name = "cluster6"
  name = "test_policy1"
  svm_name = "inter_svm"
  comment = "Comment for test policy"
}

# Example with statements including wildcard
resource "netapp-ontap_s3_policy" "full_policy" {
  cx_profile_name = "cluster6"
  name = "test_policy2"
  svm_name = "inter_svm"
  comment = "S3 policy comment with statements"
  
  statements = [
    {
      sid       = "AllowAllAccess"
      effect    = "allow"
      actions   = ["*"]
      resources = ["bucket1/*", "bucket1", "bucket2/*"]
    },
    {
      sid       = "AllowWriteAccess"
      effect    = "allow" 
      actions   = ["PutObject"]
      resources = ["bucket2/*"]
    },
    {
      sid       = "DenyAdminActions"
      effect    = "deny"
      actions   = ["DeleteBucket", "CreateBucket"]
      resources = ["*"]
    },
    {
      sid       = "testingsid"
      effect    = "deny"
      actions   = ["CreateBucket"]
      resources = ["bucket2/*"]
    }
  ]
}
