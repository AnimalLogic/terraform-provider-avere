Avere Terraform Provider
==================

Maintainers
-----------

This provider plugin was created by Liviu Constantinescu at Animal Logic, and open-sourced for maintenance & expansion by the Avere Systems team.

The current version is a beta, and under construction, so use at your own risk.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.7+
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)
-   [Python 3](https://realpython.com/installing-python/) (to run vFXT)
-   [Avere vFXT](https://pypi.org/project/vFXT/) (use `pip install vFXT` to install)

Usage
---------------------

```
provider "avere" {
  aws_access_key = (YOUR AWS ACCESS KEY)
  aws_secret_key = (YOUR AWS SECRET KEY)
  aws_deployment_region = (YOUR AWS REGION)
}
```

Building The Provider
---------------------

Install the requirements listed above. Ensure that `vFXT.py` is in your `$PATH`.

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-avere`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-avere
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-avere
$ make build
```

Copy the resulting binary to `~yourhome/.terraform.d/plugins/terraform-provider-avere`.

Using the provider
----------------------
The options of the `avere_cluster` resource are equivalent to the deployment options of the [Azure vFXT.py Script](https://github.com/Azure/AvereSDK/blob/master/docs/using_vfxt_py.md). See that repository for details on acceptable input values.

The following is an example config:

```
resource "avere_cluster" "test_cluster" {
  cluster_name = random_pet.cluster_info.id

  aws_subnet = random_pet.cluster_info.keepers.subnet_id
  aws_instance_type = random_pet.cluster_info.keepers.instance_type
  aws_security_group = random_pet.cluster_info.keepers.security_group_id

  admin_password = random_pet.cluster_info.keepers.admin_password
  node_count = random_pet.cluster_info.keepers.node_count

  core_filer_key_path = "./path/to/core_filer.key"

  use_ephemeral_storage = var.ephemeral_storage
  use_ebs_optimisation = var.ebs_optimisation
  use_at_rest_encryption = var.disk_encryption

  cache_size = random_pet.cluster_info.keepers.cache_size
  disk_iops = var.disk_iops
}
```

An example build is provided in the repository as `avere_test.tf`.

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-avere
...
```

Limitations
---------------------------

At present, this provider only supports deploying to AWS. It is also unable to determine whether the Avere cluster is still online, nor can it create a core filer or make XMLRPC calls to configure/update an existing Avere cluster.

These features will be added in future builds.