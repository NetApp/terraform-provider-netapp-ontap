package storage_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccQOSPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create qos_policy and read
			{
				Config: testAccQOSPolicyResourceBasicConfig("terraform", "tf_acc_svm", "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_qos_policy.example", "name", "terraform"),
				),
			},
			// Update a option
			{
				Config: testAccQOSPolicyResourceBasicConfig("terraform", "tf_acc_svm", "2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_qos_policy.example", "fixed.max_throughput_iops", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_qos_policy.example", "name", "terraform"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_qos_policy.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "tf_acc_qos_policy", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_qos_policy.example", "name", "tf_acc_qos_policy"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccQOSPolicyResourceBasicConfig(name string, svmName string, maxThroughputIOPS string) string {
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

resource "netapp-ontap_qos_policy" "example" {
  cx_profile_name = "cluster4"
  name = "%s"
  svm_name = "%s"
  fixed = {
    max_throughput_iops = %s
  }
}`, host, admin, password, name, svmName, maxThroughputIOPS)
}
