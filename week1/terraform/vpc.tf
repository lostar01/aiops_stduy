resource "tencentcloud_vpc" "vpc" {
  name         = "aiops-lesson"
  cidr_block   = "10.0.0.0/16"
  is_multicast = false

  tags = {
    "stack" = "aiops-lesson"
  }
}

resource "tencentcloud_subnet" "private-subnet1" {
  vpc_id            = tencentcloud_vpc.vpc.id
  name              = "private-aiops-01"
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.tencentcloud_availability_zones.default.zones.0.name
  is_multicast      = false
}

resource "tencentcloud_subnet" "private-subnet2" {
  vpc_id            = tencentcloud_vpc.vpc.id
  name              = "private-aiops-02"
  cidr_block        = "10.0.8.0/24"
  availability_zone = data.tencentcloud_availability_zones.default.zones.1.name
  is_multicast      = false
}

#public subnet
resource "tencentcloud_subnet" "public-subnet1" {
  vpc_id            = tencentcloud_vpc.vpc.id
  name              = "public-aiops-01"
  cidr_block        = "10.0.4.0/24"
  availability_zone = data.tencentcloud_availability_zones.default.zones.0.name
  is_multicast      = false
}

resource "tencentcloud_subnet" "public-subnet2" {
  vpc_id            = tencentcloud_vpc.vpc.id
  name              = "public-aiops-02"
  cidr_block        = "10.0.10.0/24"
  availability_zone = data.tencentcloud_availability_zones.default.zones.1.name
  is_multicast      = false
}


