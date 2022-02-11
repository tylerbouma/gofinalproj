package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"sync"

	"gopkg.in/yaml.v3"
)

type Virtmach struct {
	Ip            string
	Hostname      string
	Diskgb        int
	Ram           int
	Status        string
	ResourceGroup string
	Tags          []Tags
}

type Tags struct {
	TagName string `yaml:"tagName"`
	TagVal  string `yaml:"tagValue"`
}

type ResourceGroup struct {
	name string
	vms  []Virtmach
}

func (v *Virtmach) dealloc() {
	// deallocate a VM if it is running
	if v.Status != "off" {
		fmt.Println("deallocating vm:", v.Hostname)
		v.Status = "off"
	} else {
		fmt.Println(v.Hostname, "is already deallocated")
	}
}

func (v Virtmach) sysinfo() {
	// print VM info
	fmt.Print("\n--------------\n")
	fmt.Println("hostname:", v.Hostname)
	fmt.Println("ip:", v.Ip)
	fmt.Println("disk space:", v.Diskgb, "GB")
	fmt.Println("ram:", v.Ram, "GB")
	fmt.Println("VM status:", v.Status)
	fmt.Println("resource group:", v.ResourceGroup)
	for i, c := range v.Tags {
		fmt.Printf("\ntag%v name: %v\n", i, c.TagName)
		fmt.Printf("tag%v value: %v\n", i, c.TagVal)
	}
	fmt.Printf("--------------\n\n")
}

func (v *Virtmach) associate(rg *ResourceGroup) {
	// Associate VM with a group
	v.ResourceGroup = rg.name
	rg.vms = append(rg.vms, *v)
}

func (rg ResourceGroup) rginfo() {
	// print resource group information
	fmt.Print("\n--------------\n")
	fmt.Println("resource group name:", rg.name)
	fmt.Println("vms belonging to this resource group:")
	for _, r := range rg.vms {
		fmt.Println(r.Hostname)
	}
	fmt.Printf("--------------\n\n")
}

func createVM(vm Virtmach, cmd string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println(vm.Hostname, "is running", cmd)

	// uses a passed str to execute a command
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))
}

func main() {
	rg := ResourceGroup{name: "empire", vms: []Virtmach{}}

	// read a yaml config file containing VM information
	yfile, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		log.Fatal(err)
	}

	vms := make(map[string]Virtmach)

	// parse the yaml config to create new VMs
	if err = yaml.Unmarshal(yfile, &vms); err != nil {
		log.Fatal(err)
	}

	// create a random command for our "VMs" to run
	cmd := "ping -c 4 google.com"
	wg := new(sync.WaitGroup)
	for _, v := range vms {
		// create a go routine to act as a fact VM and run a simple job
		// once the job is done on all VMs allow the program to move on
		go createVM(v, cmd, wg)
		wg.Add(1)
	}

	wg.Wait()

	// deallocate a VM
	vm1 := vms["VM1"]
	vm1.dealloc()
	vm1.sysinfo()

	// associate vm1 for our resource group
	vm1.associate(&rg)
	rg.rginfo()
	vm1.sysinfo()

	// we can also point to a specific VM and it's underlying properties
	fmt.Println("grab the IP of vm2")
	fmt.Println(vms["VM2"].Ip)

	// add another vm to our resource group
	vm2 := vms["VM2"]
	vm2.associate(&rg)
	rg.rginfo()

}
