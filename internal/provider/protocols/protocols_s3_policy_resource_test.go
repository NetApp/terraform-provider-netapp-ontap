package protocols_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccProtocolsS3PolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProtocolsS3PolicyResourceConfig("tf_acc_svm"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "comment", "Test S3 policy"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "statements.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "statements.0.sid", "FullAccessToBucket1"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "statements.0.effect", "allow"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "netapp-ontap_s3_policy.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"read_only"}, // read_only is computed and may not be returned during import
				ImportStateIdFunc:       testAccProtocolsS3PolicyResourceImportStateIdFunc("netapp-ontap_s3_policy.example"),
			},
			// Update and Read testing
			{
				Config: testAccProtocolsS3PolicyResourceConfigUpdate("tf_acc_svm"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "comment", "Updated test S3 policy"),
					resource.TestCheckResourceAttr("netapp-ontap_s3_policy.example", "statements.#", "2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProtocolsS3PolicyResourceConfig(name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		return ""
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster1"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_policy" "example" {
  cx_profile_name = "cluster1"
  name           = "%s"
  svm_name       = "%s" 
  comment        = "Test S3 policy"
  
  statements = [
    {
      sid       = "FullAccessToBucket1"
      effect    = "allow"
      actions   = ["*"]
      resources = ["bucket1", "bucket1/*"]
    }
  ]
}`, host, admin, password, name, name)
}

func testAccProtocolsS3PolicyResourceConfigUpdate(name string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster1"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_s3_policy" "example" {
  cx_profile_name = "cluster1"
  name           = "%s"
  svm_name       = "%s"
  comment        = "Updated test S3 policy"
  
  statements = [
    {
      sid       = "FullAccessToBucket1"
      effect    = "allow"
      actions   = ["*"]
      resources = ["bucket1", "bucket1/*"]
    },
    {
      sid       = "DenyDeleteFromBucket2"
      effect    = "deny"
      actions   = ["DeleteObject"]
      resources = ["bucket1/*"]
    }
  ]
}`, host, admin, password, name, name)
}

func testAccProtocolsS3PolicyResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes["name"], rs.Primary.Attributes["svm_name"], rs.Primary.Attributes["cx_profile_name"]), nil
	}
}
