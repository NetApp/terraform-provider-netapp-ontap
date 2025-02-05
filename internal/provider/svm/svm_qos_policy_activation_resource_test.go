package svm_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// example ID: aeef4e4f-a663-11ef-9ca8-00a0b8bc0407
const idRegex string = "[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}"

func TestAccSvmQosPolicyActivationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSvmQosPolicyActivationResourceConfig("svm02", "performance-svm02"),
				Check: resource.ComposeTestCheckFunc(
					// Check to see if names are set correctly
					resource.TestCheckResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "svm.name", "svm02"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "qos_policy.name", "performance-svm02"),
					// Check to see if IDs are set correctly (we don't know what the value is as it changes)
					resource.TestMatchResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "svm.id", regexp.MustCompile(idRegex)),
					resource.TestMatchResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "qos_policy.id", regexp.MustCompile(idRegex)),
				),
			},
			// Update qos_policy
			{
				Config: testAccSvmQosPolicyActivationResourceConfig("svm02", "test-svm02"),
				Check: resource.ComposeTestCheckFunc(
					// Check to see if names are set correctly
					resource.TestCheckResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "svm.name", "svm02"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "qos_policy.name", "test-svm02"),
					// Check to see if IDs are set correctly (we don't know what the value is as it changes)
					resource.TestMatchResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "svm.id", regexp.MustCompile(idRegex)),
					resource.TestMatchResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "qos_policy.id", regexp.MustCompile(idRegex)),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_svm_qos_policy_activation.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "svm02", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					// Check to see if names are set correctly
					resource.TestCheckResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "svm.name", "svm02"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "qos_policy.name", "test-svm02"),
					// Check to see if IDs are set correctly (we don't know what the value is as it changes)
					resource.TestMatchResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "svm.id", regexp.MustCompile(idRegex)),
					resource.TestMatchResourceAttr("netapp-ontap_svm_qos_policy_activation.example", "qos_policy.id", regexp.MustCompile(idRegex)),
				),
			},
		},
	})
}
func testAccSvmQosPolicyActivationResourceConfig(svm, qosPolicy string) string {
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

resource "netapp-ontap_svm_qos_policy_activation" "example" {
  cx_profile_name = "cluster4"
  svm = {
	  name = "%s"
	}
	qos_policy = {
	  name = "%s"
	}
}`, host, admin, password, svm, qosPolicy)
}
