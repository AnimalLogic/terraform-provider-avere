package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"aws_access_key": {
				Type:     schema.TypeString,
				Required: true,
			},

			"aws_secret_key": {
				Type:     schema.TypeString,
				Required: true,
			},

			"aws_deployment_region": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"avere_cluster": resourceCluster(),
		},

		ConfigureFunc: providerConfigure,
	}
}

type Config struct {
	awsAccessKey        string
	awsSecretKey        string
	awsDeploymentRegion string
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		awsAccessKey:        d.Get("aws_access_key").(string),
		awsSecretKey:        d.Get("aws_secret_key").(string),
		awsDeploymentRegion: d.Get("aws_deployment_region").(string),
	}

	return &config, nil
}
