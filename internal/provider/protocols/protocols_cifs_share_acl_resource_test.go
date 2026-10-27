package protocols_test

import (
	"fmt"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsShareACLResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test missing required arguments.
			{
				Config:      testAccCifsShareACLResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// Create and read.
			// {
			// 	Config: testAccCifsShareACLResourceConfig("d_svm", "acc_test_cifs_share_acl", "BUILTIN\\Administrators", "windows", "read"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "cx_profile_name", "clustercifs"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "svm_name", "d_svm"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "share_name", "acc_test_cifs_share_acl"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "user_or_group", "BUILTIN\\Administrators"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "type", "windows"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "permission", "read"),
			// 	),
			// },
			// // Update and read.
			// {
			// 	Config: testAccCifsShareACLResourceConfig("d_svm", "acc_test_cifs_share_acl", "BUILTIN\\Administrators", "windows", "full_control"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "svm_name", "d_svm"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "share_name", "acc_test_cifs_share_acl"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "user_or_group", "BUILTIN\\Administrators"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "type", "windows"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "permission", "full_control"),
			// 	),
			// },
			// // Import and read.
			// {
			// 	ResourceName:  "netapp-ontap_cifs_share_acl.example",
			// 	ImportState:   true,
			// 	ImportStateId: fmt.Sprintf("%s,%s,%s,%s,%s", "BUILTIN\\Administrators", "windows", "acc_test_cifs_share_acl", "d_svm", "clustercifs"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "cx_profile_name", "clustercifs"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "svm_name", "d_svm"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "share_name", "acc_test_cifs_share_acl"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "user_or_group", "BUILTIN\\Administrators"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "type", "windows"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_cifs_share_acl.example", "permission", "full_control"),
			// 	),
			// },
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

// func testAccCifsShareACLResourceConfig(svmName, shareName, userOrGroup, aclType, permission string) string {
// 	host := os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
// 	admin := os.Getenv("TF_ACC_NETAPP_USER")
// 	password := os.Getenv("TF_ACC_NETAPP_PASS2")
//
// 	if host == "" || admin == "" || password == "" {
// 		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
// 		os.Exit(1)
// 	}
//
// 	return fmt.Sprintf(`
// provider "netapp-ontap" {
//   connection_profiles = [
//     {
//       name           = "clustercifs"
//       hostname       = "%s"
//       username       = "%s"
//       password       = "%s"
//       validate_certs = false
//     },
//   ]
// }
//
// resource "netapp-ontap_cifs_share_acl" "example" {
//   cx_profile_name = "clustercifs"
//   svm_name        = "%s"
//   share_name      = "%s"
//   user_or_group   = %q
//   type            = "%s"
//   permission      = "%s"
// }
// `, host, admin, password, svmName, shareName, userOrGroup, aclType, permission)
// }
