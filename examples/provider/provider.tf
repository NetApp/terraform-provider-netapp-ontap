terraform {
  required_providers {
    netapp-ontap = {
      source = "NetApp/netapp-ontap"
      version = "0.0.1"
    }
  }
}


provider "netapp-ontap" {
  # A connection profile defines how to interface with an ONTAP cluster or svm.
  # At least one is required.
  connection_profiles = [
    {
      name = "cluster1"
      hostname = "********219"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster2"
      hostname = "********222"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster3"
      hostname = "10.193.176.159"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster4"
      hostname = "10.193.180.108"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "clustercifs"
      hostname = "10.193.73.189"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    }
  ]
}
