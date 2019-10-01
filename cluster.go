package main

import (
	"bufio"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"admin_password": {
				Type:     schema.TypeString,
				Required: true,
			},

			"aws_subnet": {
				Type:     schema.TypeString,
				Required: true,
			},

			"aws_instance_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "r3.2xlarge",
			},

			"aws_security_group": {
				Type:     schema.TypeString,
				Required: true,
			},

			"core_filer_key_path": {
				Type:     schema.TypeString,
				Required: true,
			},

			"node_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000,
			},

			"node_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},

			"disk_iops": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"use_ephemeral_storage": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"use_ebs_optimisation": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"use_at_rest_encryption": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"management_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// The following four methods are used by Terraform to capture and control resource state.
func resourceClusterCreate(d *schema.ResourceData, m interface{}) error {
	runVFXTClusterCreateScript(d, m)
	masterIP := d.Get("management_address").(string)
	d.SetId(masterIP)

	return resourceClusterRead(d, m)
}

func resourceClusterRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceClusterUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceClusterRead(d, m)
}

func resourceClusterDelete(d *schema.ResourceData, m interface{}) error {
	runVFXTClusterDeleteScript(d, m)

	return nil
}

// Helper function to check that vFXT.py is installed.
func checkvFXTExists() {
	_, err := exec.LookPath("vFXT.py")
	if err != nil {
		log.Fatal("vFXY.py is not installed, or not in your PATH\n")
	}
}

// vFXT script invocation for creating a cluster
func runVFXTClusterCreateScript(d *schema.ResourceData, m interface{}) {
	checkvFXTExists()

	opt := func(key string) string {
		if str, ok := d.Get(key).(string); ok {
			return str
		} else {
			return strconv.Itoa(d.Get(key).(int))
		}
	}

	ebsOpts, ephemeralOpts, encryptionOpts, diskOpts := PrepareVariables(d)

	args := []string{"--create", "--cloud-type", "aws", "--cluster-name",
		opt("cluster_name"), "--access-key", m.(*Config).awsAccessKey, "--secret-key", m.(*Config).awsSecretKey,
		"--subnet", opt("aws_subnet"), "--region", m.(*Config).awsDeploymentRegion,
		"--admin-password", opt("admin_password"), "--node-cache-size", opt("node_size"),
		"--nodes", opt("node_count"), "--instance-type", opt("aws_instance_type"),
		"--security-group", opt("aws_security_group"),
		"--core-filer-key-file", d.Get("core_filer_key_path").(string),
		"--no-corefiler", "--debug"}

	if ephemeralOpts != "" {
		args = append(args, ephemeralOpts)
	}
	if ebsOpts != "" {
		args = append(args, ebsOpts)
	}
	if encryptionOpts != "" {
		args = append(args, encryptionOpts)
	}
	if diskOpts != "" {
		args = append(args, diskOpts)
	}

	vFXTCreateCommand := exec.Command("vFXT.py", args...)

	runAndEchoConsoleOutput(vFXTCreateCommand, d)
}

// vFXT script invocation for deleting a cluster
func runVFXTClusterDeleteScript(d *schema.ResourceData, m interface{}) {
	checkvFXTExists()

	opt := func(key string) string {
		if str, ok := d.Get(key).(string); ok {
			return str
		} else {
			return strconv.Itoa(d.Get(key).(int))
		}
	}

	args := []string{"--destroy", "--cloud-type", "aws",
		"--access-key", m.(*Config).awsAccessKey, "--secret-key", m.(*Config).awsSecretKey,
		"--subnet", opt("aws_subnet"), "--region", m.(*Config).awsDeploymentRegion,
		"--admin-password", opt("admin_password"), "--management-address", opt("management_address")}

	vFXTDeleteCommand := exec.Command("vFXT.py", args...)

	runAndEchoConsoleOutput(vFXTDeleteCommand, d)
}

// Prepares some command line option strings and convenience methods for the vFXT function invocation
func PrepareVariables(d *schema.ResourceData) (string, string, string, string) {
	ephemeral, ephemeralOk := d.GetOk("use_ephemeral_storage")
	ebs, ebsOk := d.GetOk("use_ebs_optimisation")
	encryption, encryptionOk := d.GetOk("use_at_rest_encryption")
	disk, diskOk := d.GetOk("disk_iops")
	var ebsOpts string
	if !ebsOk && ephemeralOk {
		ebs = !ephemeral.(bool)
		ebsOk = true
	}
	_ = parseOpt(ebsOk, ebs, &ebsOpts, "", "--no-ebs-optimized")
	var ephemeralOpts string
	_ = parseOpt(ephemeralOk, ephemeral, &ephemeralOpts, "--ephemeral", "")
	var encryptionOpts string
	_ = parseOpt(encryptionOk, encryption, &encryptionOpts, "", "--no-disk-encryption")
	var diskOpts string
	if !diskOk {
		diskOpts = ""
	} else {
		diskOpts = "--data-disk-iops " + disk.(string)
	}

	return ebsOpts, ephemeralOpts, encryptionOpts, diskOpts
}

// Runs a vFXT command, captures the management address, and echoes all output to Terraform via the log function
func runAndEchoConsoleOutput(cmd *exec.Cmd, d *schema.ResourceData) {
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	terraformRd, terraformWr, pipeErr := os.Pipe()
	if pipeErr != nil {
		log.Fatalf("Can't create out-pipe to Terraform.")
	}

	go func() {
		scanner := bufio.NewScanner(terraformRd)
		for scanner.Scan() {
			stdoutLine := scanner.Text()
			log.Printf("vFXT.py | %s", stdoutLine)

			captureManagementIPIfEncountered(stdoutLine, d)
		}
	}()

	var errStdout, errStderr error
	stdout := io.MultiWriter(terraformWr, os.Stdout)
	stderr := io.MultiWriter(terraformWr, os.Stderr)

	err := cmd.Start()
	if err != nil {
		log.Fatalf("vFXT.py failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("vFXT.py failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture vFXT.py stdout or stderr\n")
	}
}

func captureManagementIPIfEncountered(stdoutLine string, d *schema.ResourceData) {
	var managementIPParser = regexp.MustCompile(`^address=(\d+\.\d+\.\d+\.\d+)$`)
	managementIP := managementIPParser.FindStringSubmatch(stdoutLine)
	if managementIP != nil {
		log.Printf("Captured and saved to Terraform state... Management IP: %s", managementIP[1])

		if err := d.Set("management_address", managementIP[1]); err != nil {
			_ = fmt.Errorf("error setting cluster management address: %s", err)
		}
	}
}

// Helper function for parameter opts
func parseOpt(optOk bool, opt interface{}, optOpts *string, okAndTrueString string, falseString string) error {
	if optOk && opt.(bool) {
		*optOpts = okAndTrueString
	} else {
		*optOpts = falseString
	}

	return nil
}
