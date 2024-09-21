# Get availability zones
data "tencentcloud_availability_zones" "default" {}

# Get availability images
data "tencentcloud_images" "default" {
  image_type = ["PUBLIC_IMAGE"]
  os_name    = "ubuntu"
}

# Get availability instance types
data "tencentcloud_instance_types" "default" {
  # 机型族
  filter {
    name   = "instance-family"
    values = ["SA5"]
  }

  cpu_core_count = 2
  memory_size    = 4
  exclude_sold_out = true
}

# Create a cvm
resource "tencentcloud_instance" "aiops-cvm1" {
  instance_name              = "aiops-lesson1"
  availability_zone          = data.tencentcloud_availability_zones.default.zones.0.name
  image_id                   = data.tencentcloud_images.default.images.0.image_id
  instance_type              = data.tencentcloud_instance_types.default.instance_types.0.instance_type
  system_disk_type           = "CLOUD_BSSD"
  system_disk_size           = 20
  allocate_public_ip         = true
  internet_max_bandwidth_out = 10
  #security_groups            = [tencentcloud_security_group.aiops-cvm1.id]
  orderly_security_groups    = [tencentcloud_security_group.aiops-cvm1.id]
  vpc_id                     = tencentcloud_vpc.vpc.id
  subnet_id                  = tencentcloud_subnet.public-subnet1.id
  instance_charge_type       = "SPOTPAID"
  spot_instance_type         = "ONE-TIME"
  spot_max_price             = 0.1
  key_ids                    = ["skey-gyrtffml"]
  user_data                  = base64encode(file("./userdata.sh"))
 
  tags = {
    stack = "aiops"
  }
}
