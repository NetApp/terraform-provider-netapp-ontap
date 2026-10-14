package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsShareACLResource(t *testing.T) {
	const (
		resourceName = "netapp-ontap_cifs_share_acl.example"
		svmName      = "one"
		shareName    = "acc_test_cifs_share_acl"
		userOrGroup  = "BUILTIN\\Administrators"
		aclType      = "windows"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test missing required arguments.
			{
				Config:      testAccCifsShareACLResourceConfigMissingVars(svmName),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// Create and read.
			{
				Config: testAccCifsShareACLResourceConfig(svmName, shareName, userOrGroup, aclType, "read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cx_profile_name", "clustercifs"),
					resource.TestCheckResourceAttr(resourceName, "svm_name", svmName),
					resource.TestCheckResourceAttr(resourceName, "share_name", shareName),
					resource.TestCheckResourceAttr(resourceName, "user_or_group", userOrGroup),
					resource.TestCheckResourceAttr(resourceName, "type", aclType),
					resource.TestCheckResourceAttr(resourceName, "permission", "read"),
				),
			},
			// Update and read.
			{
				Config: testAccCifsShareACLResourceConfig(svmName, shareName, userOrGroup, aclType, "full_control"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "svm_name", svmName),
					resource.TestCheckResourceAttr(resourceName, "share_name", shareName),
					resource.TestCheckResourceAttr(resourceName, "user_or_group", userOrGroup),
					resource.TestCheckResourceAttr(resourceName, "type", aclType),
					resource.TestCheckResourceAttr(resourceName, "permission", "full_control"),
				),
			},
			// Import and read.
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s,%s", userOrGroup, aclType, shareName, svmName, "clustercifs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cx_profile_name", "clustercifs"),
					resource.TestCheckResourceAttr(resourceName, "svm_name", svmName),
					resource.TestCheckResourceAttr(resourceName, "share_name", shareName),
					resource.TestCheckResourceAttr(resourceName, "user_or_group", userOrGroup),
					resource.TestCheckResourceAttr(resourceName, "type", aclType),
					resource.TestCheckResourceAttr(resourceName, "permission", "full_control"),
				),
			},
		},
	})
}

func testAccCifsShareACLResourceConfigMissingVars(svmName string) string {
	return fmt.Sprintf(`
resource "netapp-ontap_cifs_share_acl" "example1" {
	  svm_name = "%s"
	  share_name = "testme"
	  user_or_group = "Everyone"
	  type = "windows"
	  permission = "read"
}
`, svmName)
}

func testAccCifsShareACLResourceConfig(svmName, shareName, userOrGroup, aclType, permission string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
		os.Exit(1)
	}

	return fmt.Sprintf(`
provider "netapp-ontap" {
  connection_profiles = [
    {
      name           = "clustercifs"
      hostname       = "%s"
      username       = "%s"
      password       = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_cifs_share" "acl_test" {
  cx_profile_name = "clustercifs"
  name            = "%s"
  svm_name        = "%s"
  path            = "/"
}

resource "netapp-ontap_cifs_share_acl" "example" {
  cx_profile_name = "clustercifs"
  svm_name        = "%s"
  share_name      = netapp-ontap_cifs_share.acl_test.name
  user_or_group   = %q
  type            = "%s"
  permission      = "%s"
}
`, host, admin, password, shareName, svmName, svmName, userOrGroup, aclType, permission)
}
