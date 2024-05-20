package protocols_test

import (
	"fmt"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsLocalGroupMemberResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCifsLocalGroupMemberResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// create with basic argument on a local group member
			{
				// configuration of group name and memmber name have to be double \ otherwise it will be treated as escape character
				Config: testAccCifsLocalGroupMemberResourceConfig("svm3", "SVM3_SERVER\\\\accgroup1", "SVM3_SERVER\\\\accuser3"),
				Check: resource.ComposeTestCheckFunc(
					// check member
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "member", "SVM3_SERVER\\accuser3"),
					// check group_name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "group_name", "SVM3_SERVER\\accgroup1"),
					// check svm_name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "svm_name", "svm3"),
					// check ID
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "id"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_protocols_cifs_local_group_member_resource.example1",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "SVM3_SERVER\\accuser3", "SVM3_SERVER\\accgroup1", "svm3", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "svm_name", "svm3"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "group_name", "SVM3_SERVER\\accgroup1"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "member", "SVM3_SERVER\\accuser3"),
					// check id
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_group_member_resource.example1", "id"),
				),
			},
		},
	})
}

func testAccCifsLocalGroupMemberResourceConfigMissingVars(svmName string) string {
	return fmt.Sprintf(`
resource "netapp-ontap_protocols_cifs_local_group_member_resource" "example1" {
	  svm_name = "%s"
	  group_name = "SVM3_SERVER\\accgroup1"
	  member = "SVM3_SERVER\\accuser3"
}
`, svmName)
}

func testAccCifsLocalGroupMemberResourceConfig(svmName, groupName, member string) string {
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
resource "netapp-ontap_protocols_cifs_local_group_member_resource" "example1" {
	cx_profile_name = "cluster4"
	  svm_name = "%s"
	  group_name = "%s"
	  member = "%s"
}
`, host, admin, password, svmName, groupName, member)
}
