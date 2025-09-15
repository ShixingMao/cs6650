# You probably want to keep your ip address a secret as well
variable "ssh_cidr" {
  type        = string
  description = "Your home IP in CIDR notation"
}

# name of the existing AWS key pair
variable "ssh_key_name" {
  type        = string
  description = "Name of your existing AWS key pair"
}

# The provider of your cloud service, in this case it is AWS.
provider "aws" {
  region = "us-west-2" # Which region you are working on
}

# Your ec2 instances - NOW CREATING 2 INSTANCES!
resource "aws_instance" "demo-instance" {
  count                  = 2  # CREATE 2 INSTANCES
  ami                    = data.aws_ami.al2023.id
  instance_type          = "t2.micro"
  iam_instance_profile   = "LabInstanceProfile"
  vpc_security_group_ids = [aws_security_group.ssh.id]
  key_name               = var.ssh_key_name

  tags = {
    Name = "terraform-instance-${count.index + 1}"  # Will name them terraform-instance-1 and terraform-instance-2
  }
}

# Your security that grants ssh access from
# your ip address to your ec2 instance
# ALSO NOW INCLUDES PORT 8080 FOR YOUR APP!
resource "aws_security_group" "ssh" {
  name        = "allow_ssh_and_app"
  description = "SSH and app access"
  
  # SSH access
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.ssh_cidr]
  }
  
  # App port 8080 - ADDED THIS!
  ingress {
    description = "App Port"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Allow from anywhere, or use [var.ssh_cidr] to restrict to your IP
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# latest Amazon Linux 2023 AMI
data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }
}

# UPDATED OUTPUTS FOR BOTH INSTANCES
output "ec2_public_dns" {
  value = {
    instance_1 = aws_instance.demo-instance[0].public_dns
    instance_2 = aws_instance.demo-instance[1].public_dns
  }
}

output "ec2_public_ip" {
  value = {
    instance_1 = aws_instance.demo-instance[0].public_ip
    instance_2 = aws_instance.demo-instance[1].public_ip
  }
}