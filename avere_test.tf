variable "aws_access_key" {
  type        = string
  description = "AWS CLI access key for Avere user"
}

variable "aws_secret_key" {
  type        = string
  description = "AWS CLI secret key for Avere user."
}

variable "instance_type" {
  type        = string
  description = "The type of instance to use for the Avere cluster."
}

variable "node_size" {
  type        = number
  description = "The storage size of each Avere cluster node, in GB."
}

variable "node_count" {
  type        = number
  description = "The number of desired nodes in the Avere cluster."
}

variable "security_group_id" {
  type        = string
  description = "The ID for the security group to use for Avere nodes."
}

variable "subnet_id" {
  type        = string
  description = "The id of the AWS subnet you wish to deploy the Avere cluster in."
}

variable "ephemeral_storage" {
  type        = bool
  default     = true
  description = "If true, Avere caching cluster will use ephemeral storage on its nodes, instead of mounted EBS volumes."
}

variable "ebs_optimisation" {
  type        = bool
  description = "If true, the storage caching cluster will use EBS Optimisation on its instances."
}

variable "disk_encryption" {
  type        = bool
  default     = true
  description = "If true, at-rest encryption will be used for data in the storage cache."
}

variable "deployment_region" {
  type        = string
  description = "The AWS region to deploy in."
}

resource "random_pet" "cluster_info" {
  keepers = {
    subnet_id         = var.subnet_id
    node_size         = var.node_size
    instance_type     = var.instance_type
    deployment_region = var.deployment_region
    node_count        = var.node_count
    //noinspection HILUnresolvedReference
    admin_password    = random_password.password.result
    security_group_id = var.security_group_id
  }
}

resource "random_password" "password" {
  length  = 16
  special = false
}

provider "avere" {
  aws_access_key        = var.aws_access_key
  aws_secret_key        = var.aws_secret_key
  aws_deployment_region = random_pet.cluster_info.keepers.deployment_region
}

resource "local_file" "filer_key" {
  content  = ""
  filename = "./test-filer.key"
}

resource "avere_cluster" "test_cluster" {
  cluster_name = join("-", ["test-cluster", random_pet.cluster_info.id])

  aws_subnet         = random_pet.cluster_info.keepers.subnet_id
  aws_instance_type  = random_pet.cluster_info.keepers.instance_type
  aws_security_group = random_pet.cluster_info.keepers.security_group_id

  admin_password = random_pet.cluster_info.keepers.admin_password
  node_count     = random_pet.cluster_info.keepers.node_count

  core_filer_key_path = local_file.filer_key.filename

  use_ephemeral_storage  = var.ephemeral_storage
  use_ebs_optimisation   = var.ebs_optimisation
  use_at_rest_encryption = var.disk_encryption

  node_size = random_pet.cluster_info.keepers.node_size
}

output "admin_password" {
  //noinspection HILUnresolvedReference
  value      = random_password.password.result
  depends_on = [avere_cluster.test_cluster]
}