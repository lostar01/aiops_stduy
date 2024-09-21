# Create security group
resource "tencentcloud_security_group" "aiops-cvm1" {
  name        = "aiops-lesson1-sg"
  description = "make it accessible for both production and stage ports"
}

# Create security group rule allow ssh request
resource "tencentcloud_security_group_rule" "ssh" {
  security_group_id = tencentcloud_security_group.aiops-cvm1.id
  type              = "ingress"
  cidr_ip           = "0.0.0.0/0"
  ip_protocol       = "tcp"
  port_range        = "22"
  policy            = "accept"
}

resource "tencentcloud_security_group_rule" "k3s" {
  security_group_id = tencentcloud_security_group.aiops-cvm1.id
  type              = "ingress"
  cidr_ip           = "0.0.0.0/0"
  ip_protocol       = "tcp"
  port_range        = "6443"
  policy            = "accept"
}

# Egress
resource "tencentcloud_security_group_rule" "egress_443" {
  security_group_id = tencentcloud_security_group.aiops-cvm1.id
  type              = "egress"
  cidr_ip           = "0.0.0.0/0"
  ip_protocol       = "tcp"
  port_range        = "443"
  policy            = "accept"
}

resource "tencentcloud_security_group_rule" "egress_80" {
  security_group_id = tencentcloud_security_group.aiops-cvm1.id
  type              = "egress"
  cidr_ip           = "0.0.0.0/0"
  ip_protocol       = "tcp"
  port_range        = "80"
  policy            = "accept"
}

