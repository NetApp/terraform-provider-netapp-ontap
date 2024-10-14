package cluster_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClusterScheduleResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test create interval schedule error
			{
				Config:      testAccClusterScheduleResourceIntervalConfigAlias("non-existant", "wrongvalue"),
				ExpectError: regexp.MustCompile("error creating cluster_schedule"),
			},
			// Create interval and read
			{
				Config: testAccClusterScheduleResourceIntervalConfigAlias("tf-interval-schedule-test", "PT8M30S"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.example", "name", "tf-interval-schedule-test"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.example", "interval", "PT8M30S"),
				),
			},
			// update and read
			{
				Config: testAccClusterScheduleResourceIntervalConfigAlias("tf-interval-schedule-test", "PT8M20S"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.example", "name", "tf-interval-schedule-test"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.example", "interval", "PT8M20S"),
				),
			},
			// Test importing a interval job schedule resource
			{
				ResourceName:  "netapp-ontap_cluster_schedule.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf-interval-schedule-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.example", "name", "tf-interval-schedule-test"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.example", "interval", "PT8M20S"),
				),
			},
			// Create cron schedule and read
			{
				Config: testAccClusterScheduleCreateResourceCronConfigAlias("tf-cron-schedule-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.cron-example", "name", "tf-cron-schedule-test"),
				),
			},
			// Update cron schedule and read
			{
				Config: testAccClusterScheduleUpdateResourceCronConfigAlias("tf-cron-schedule-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.cron-example", "name", "tf-cron-schedule-test"),
				),
			},
			// Test importing a cron job schedule resource
			{
				ResourceName:  "netapp-ontap_cluster_schedule.cron-example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf-cron-schedule-test", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.cron-example", "name", "tf-cron-schedule-test"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.cron-example", "cron.days.0", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_cluster_schedule.cron-example", "cron.weekdays.2", "4"),
				),
			},
		},
	})
}

func testAccClusterScheduleResourceIntervalConfigAlias(name string, interval string) string {
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

resource "netapp-ontap_cluster_schedule_resource" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  interval = "%s"
}`, host, admin, password, name, interval)
}

func testAccClusterScheduleCreateResourceCronConfigAlias(name string) string {
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

resource "netapp-ontap_cluster_schedule_resource" "cron-example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  cron = {
    minutes = [1, 2, 3, 4]
    hours = [10]
    days = [1, 2]
    months = [6, 7]
    weekdays = [1, 3, 4]
  }
}`, host, admin, password, name)
}

func testAccClusterScheduleUpdateResourceCronConfigAlias(name string) string {
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

resource "netapp-ontap_cluster_schedule_resource" "cron-example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "%s"
  cron = {
    minutes = [4, 5, 6]
    hours = [2, 3]
    days = [2]
    months = [1, 6, 7]
    weekdays = [3, 4, 5]
  }
}`, host, admin, password, name)
}
