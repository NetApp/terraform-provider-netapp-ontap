package cluster_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClusterPeerResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create svm peer and read
			{
				Config: testAccClusterPeerResourceConfigAlias("10.193.180.110", "10.193.176.189"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_peer.example", "remote.ip_addresses.0", "10.193.180.110"),
				),
			},
			// Update applications
			{
				Config: testAccClusterPeerResourceConfigAlias("10.193.180.109", "10.193.176.189"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_peer.example", "remote.ip_addresses.0", "10.193.180.109"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_cluster_peer.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "acc_test_cluster2", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_peer.example", "name", "acc_test_cluster2"),
				),
			},
		},
	})
}
func testAccClusterPeerResourceConfigAlias(remotIP, sourceIP string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST2")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	password2 := os.Getenv("TF_ACC_NETAPP_PASS2")
	host2 := os.Getenv("TF_ACC_NETAPP_HOST")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST2, TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, TF_ACC_NETAPP_PASS and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
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
}`, host, admin, password2, host2, admin, password, remotIP, sourceIP)
}
