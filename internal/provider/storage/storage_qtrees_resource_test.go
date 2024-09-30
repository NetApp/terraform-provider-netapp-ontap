package storage_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageQtreesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test create the resource
			{
				Config: createQtree("acc_test_qtree", "temp_root", "temp"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_qtrees.example", "name", "acc_test_qtree"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_qtrees.example", "user.name", "nobody"),
				),
			},
			{
				Config: updateGroupAndUser("acc_test_qtree", "temp_root", "temp"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_qtrees.example", "name", "acc_test_qtree"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_qtrees.example", "user.name", "root"),
					resource.TestCheckResourceAttr("netapp-ontap_storage_qtrees.example", "group.name", "root"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_storage_qtrees.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "acc_import", "terraform_root", "terraform", "cluster5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_storage_qtrees.example", "name", "accFlexcache"),
				),
			},
		},
	})
}

func createQtree(name, volName, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_storage_qtrees" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  volume_name = "%s"
  svm_name = "%s"
  security_style = "unix" 
  user = {
    name = "nobody"
  }
  group = {
    name = "nobody"
  }
}`, host, admin, password, name, volName, svmName)
}

func updateGroupAndUser(name, volName, svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")

	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster5"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_storage_qtrees" "example" {
  cx_profile_name = "cluster5"
  name = "%s"
  volume_name = "%s"
  svm_name = "%s"
  security_style = "unix" 
  user = {
    name = "root"
  }
  group = {
    name = "root"
  }
}`, host, admin, password, name, volName, svmName)
}
