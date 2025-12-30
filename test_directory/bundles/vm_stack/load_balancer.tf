resource "vsphere_virtual_machine" "load_balancer" {
  name             = "load-balancer"
  resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
  datastore_id     = data.vsphere_datastore.datastore.id
  num_cpus         = 2
  memory           = 2048
  guest_id         = "otherLinux64Guest"

  network_interface {
    network_id = data.vsphere_network.network.id
  }

  disk {
    label = "disk0"
    size  = 40
  }
}
