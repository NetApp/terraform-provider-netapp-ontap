package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsUserGroupPrivilegeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCifsUserGroupPrivilegeResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// create with basic argument on a local user
			{
				Config: testAccCifsUserGroupPrivilegeResourceConfig("svm3", "accuser1", "sechangenotifyprivilege"),
				Check: resource.ComposeTestCheckFunc(
					// check name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "name", "accuser1"),
					// check svm_name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "svm_name", "svm3"),
					// check ID
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "id"),
					// check privileges
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "privileges.*", "sechangenotifyprivilege"),
				),
			},
			// update one privilege
			{
				Config: testAccCifsUserGroupPrivilegeResourceConfig("svm3", "accuser1", "setakeownershipprivilege"),
				Check: resource.ComposeTestCheckFunc(
					// check user name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "name", "accuser1"),
					// check id
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "id"),
					// check updated privileges
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "privileges.*", "setakeownershipprivilege"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_protocols_cifs_user_group_privilege_resource.example1",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "accuser1", "svm3", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "svm_name", "svm3"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "name", "accuser1"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "privileges.*", "sesecurityprivilege"),
					// check id
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_user_group_privilege_resource.example1", "id"),
				),
			},
		},
	})
}

func testAccCifsUserGroupPrivilegeResourceConfigMissingVars(svmName string) string {
	return fmt.Sprintf(`
	resource "netapp-ontap_protocols_cifs_user_group_privilege_resource" "example1" {
		svm_name = "%s"
	}
	`, svmName)
}

func testAccCifsUserGroupPrivilegeResourceConfig(svmName string, name string, privilege string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_protocols_cifs_user_group_privilege_resource" "example1" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	privileges = ["%s", "sesecurityprivilege"]
}`, host, admin, password, svmName, name, privilege)
}
