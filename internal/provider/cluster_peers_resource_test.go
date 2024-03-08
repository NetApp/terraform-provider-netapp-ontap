package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClusterPeersResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create svm peer and read
			{
				Config: testAccClusterPeersResourceConfig("10.193.180.110", "10.193.176.189"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_peers_resource.example", "remote.ip_addresses.0", "10.193.180.110"),
				),
			},
			// Update applications
			{
				Config: testAccClusterPeersResourceConfig("10.193.180.109", "10.193.176.189"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_peers_resource.example", "remote.ip_addresses.0", "10.193.180.109"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_cluster_peers_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "vinaykuscluster-1", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_peers_resource.example", "name", "vinaykuscluster-1"),
				),
			},
		},
	})
}
func testAccClusterPeersResourceConfig(remote_ip, source_ip string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	host2 := os.Getenv("TF_ACC_NETAPP_HOST")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST2, TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
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
	{
		name = "cluster3"
		hostname = "%s"
		username = "%s"
		password = "%s"
		validate_certs = false
	},
  ]
}

resource "netapp-ontap_cluster_peers_resource" "example" {
  cx_profile_name = "cluster4"
  remote = {
    ip_addresses = ["%s"]
  }
  source_details = {
    ip_addresses = ["%s"]
  }
  peer_cx_profile_name = "cluster3"
  passphrase = "12345678"
  peer_applications = ["snapmirror"]
}`, host, admin, password, host2, admin, password, remote_ip, source_ip)
}
