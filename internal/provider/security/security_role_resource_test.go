package security_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
)

func TestAccSecurityRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityRoleResourceConfig("acc_test_security_role"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_role.security_role", "name", "acc_test_security_role"),
				),
			},
			// Test adding a new  priviledge to the security role
			{
				Config: AddPriviledge("acc_test_security_role"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_role.security_role", "name", "acc_test_security_role"),
				),
			},
			// Test editing a priviledge and deleting a priviledge from the security role
			{
				Config: EditPriviledgeAndDeletePriviledge("acc_test_security_role"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_role.security_role", "name", "acc_test_security_role"),
				),
			},
			// Test importing a security role
			{
				ResourceName:  "netapp-ontap_security_role.security_role",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "vsadmin", "tf_acc_svm", "cluster2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_security_role.security_role", "name", "vsadmin"),
				),
			},
		},
	})
}

func testAccSecurityRoleResourceConfig(name string) string {
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
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_role" "security_role" {
	# required to know which system to interface with
	cx_profile_name = "cluster2"
	name = "%s"
	svm_name = "tf_acc_svm"
	privileges = [
	  {
	  access = "all"
	  path = "lun"
	}
	]
  }
`, host, admin, password, name)
}

func AddPriviledge(name string) string {
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
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_role" "security_role" {
	# required to know which system to interface with
	cx_profile_name = "cluster2"
	name = "%s"
	svm_name = "tf_acc_svm"
	privileges = [
	  {
	  access = "all"
	  path = "lun"
	},
	{
	  access = "all"
	  path = "vserver"
	  query = "-vserver acc_test"
	}
	]
  }
`, host, admin, password, name)
}

func EditPriviledgeAndDeletePriviledge(name string) string {
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
      name = "cluster2"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_security_role" "security_role" {
	# required to know which system to interface with
	cx_profile_name = "cluster2"
	name = "%s"
	svm_name = "tf_acc_svm"
	privileges = [
	{
	  access = "all"
	  path = "vserver"
	  query = "-vserver tf_acc_svm|temp"
	}
	]
  }
`, host, admin, password, name)
}
