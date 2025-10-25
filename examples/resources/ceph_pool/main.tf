resource "ceph_pool" "example" {
  name        = "example-pool"
  pool_type   = "replicated"
  pg_num      = 128
  pgp_num     = 128
  size        = 3
  application = "rbd"
}