package main

// ################## imports
import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ################## global variables

const appVersion string = "0.1"

// filename with xml meta data information
var exportFile string = ""

// xe binary which is used to import the metadata
var xeBinary string = "/usr/bin/xe"

// VirtualMachines contain none or many VirtualMachine.
type VirtualMachines struct {
	XMLName xml.Name         `xml:"vms"`
	Vms     []VirtualMachine `xml:"vm"`
}

// VirtualMachine represents a vm
type VirtualMachine struct {
	XMLName   xml.Name  `xml:"vm"`
	NameLabel string    `xml:"name,attr"`
	UUID      string    `xml:"uuid,attr"`
	Ps        Parents   `xml:"parents"`
	VBDs      VBDS      `xml:"vbds"`
	Snaps     Snapshots `xml:"snapshots"`
}

// Parents of a vm as list
type Parents struct {
	XMLName xml.Name `xml:"parents"`
	Ps      []Parent `xml:"parent"`
}

// Parent of a vm, part of the list
type Parent struct {
	XMLName    xml.Name `xml:"parent"`
	UUID       string   `xml:"uuid,attr"`
	Selfparent bool     `xml:"selfparent,attr"`
}

// VBDS of a vm as list
type VBDS struct {
	XMLName xml.Name `xml:"vbds"`
	Vbds    []VBD    `xml:"vbd"`
}

// VBD of a vm, part of the list
type VBD struct {
	XMLName      xml.Name `xml:"vbd"`
	UUID         string   `xml:"uuid,attr"`
	VbdType      string   `xml:"type,attr"`
	VdiNameLabel string   `xml:"vdi-name-label,attr"`
}

// Snapshots of a vm as list.
type Snapshots struct {
	XMLName xml.Name   `xml:"snapshots"`
	Snaps   []Snapshot `xml:"snapshot"`
}

// Snapshot of a vm, part of the list.
type Snapshot struct {
	XMLName         xml.Name `xml:"snapshot"`
	UUID            string   `xml:"uuid,atrr"`
	NameLable       string   `xml:"name-lable,atrr"`
	NameDescription string   `xml:"name-description,atrr"`
	IsVmssSnapshot  string   `xml:"is-vmss-snapshot,atrr"`
}

// ############################# FUNCTIONS

// parseCommandOptions is the capsuled function for checking
// the command line parameter.
func parseCommandOptions() {
	//flag.StringVar(&xeBinary, "xebinary", xeBinary, "Absolute path to xe binary including executable.")

	flag.StringVar(&exportFile, "outfile", exportFile, "Filename with meta data to export.")

	// The flag package provides a default help printer via -h switch
	var versionFlag *bool = flag.Bool("v", false, "Print the version number.")

	flag.Parse() // Scan the arguments list

	// check if flag -v is given, print out version end exit.
	if *versionFlag {
		fmt.Println("Version:", appVersion)
		os.Exit(0)
	}

	// check if an outfile is specified.
	if len(exportFile) == 0 {
		exportFile = "vms.export.example.xml"
		log.Fatal("No export file given. Exiting...")
	}
}

// getVMs returns an array of all uuid of vms via xe command line.
// The set of vms can be filtered by xe command line filter, e.g. power-state=running.
// For filter, please consult the xe command line reference:
//  https://docs.citrix.com/en-us/citrix-hypervisor/command-line-interface.html
func getVMs(xefilter string) []string {
	// validate xe filter expression
	if !validateXeFilter(xefilter) {
		log.Fatal("Wrong filter specified. Exiting...")
	}

	// run xe vm-list --minimal to get all uuids of vms in a comma separated list
	cmd := exec.Command(xeBinary, "vm-list", "--minimal")
	// buffer to read cmd output
	var out bytes.Buffer
	//assign cmd stdout to buffer
	cmd.Stdout = &out

	// check command for errors
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	//return array of uuids of all vms
	return strings.Split(string(out.Bytes()), ",")
}

// validateXeFilter does a rudimentary validation of the given
// filter as string for the xe command and returns true if
// filter is correct otherwise false.
//
// Actually only the following filter are supported:
// * power-state=runningreturn getVmObjectsAsUuidList(uuid, "params=snapshots")
func validateXeFilter(filter2validate string) bool {
	switch filter2validate {
	case "power-state=running":
		log.Printf("Using filter %v", filter2validate)
		return true
	default:
		log.Printf("No filter defined, using none.")
	}
	return false
}

// getVMAttribute gets the attributes of a vm by the specified uuid.
//
// Result of this function is an array of strings which contain one of the following:
// * uuids of snapshots or vbds, when parameter params is parent or snapshot, e.g. params=snapshots
// * atrribute value, when parameter params is one concrete attribute, e.g. params=name-description
func getVMAttribute(uuid string, params string) []string {

	// run xe vm-list --minimal to get all uuids of vms in a comma separated list
	cmd := exec.Command(xeBinary, "vm-list", "uuid="+uuid, params, "--minimal")
	// buffer to read cmd output
	var out bytes.Buffer
	//assign cmd stdout to buffer
	cmd.Stdout = &out

	// check command for errors
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	//return array of uuids of all vms
	return strings.Split(string(out.Bytes()), ",")
}

// getVMDetails returns the vm informations as map of attributes
// The attribute name-label is got by function getVmObjectsAsUuidList
func getVMDetails(uuid string) map[string]string {
	var vm map[string]string
	vm = make(map[string]string)
	vm["uuid"] = uuid
	vm["name-label"] = getVMAttribute(uuid, "params=name-label")[0]
	return vm
}

// getVMSnapshots is getting all uuids of snapshots for vm with uuid.
// Result is an array of all snapshot uuids.
func getVMSnapshots(uuid string) []string {
	return getVMAttribute(uuid, "params=snapshots")
}

// getSnapshotDetails gets the value of an snapshot with the specified uuid.
//
// All attributes with its values were returned as map of strings.
func getSnapshotDetails(uuid string) map[string]string {
	var s map[string]string
	s = make(map[string]string)
	s["uuid"] = uuid
	s["name-lable"] = getSnapshotAttribute(uuid, "params=name-lable")[0]
	s["name-description"] = getSnapshotAttribute(uuid, "params=name-description")[0]
	s["is-vmss-snapshot"] = getSnapshotAttribute(uuid, "params=is-vmss-snapshot")[0]
	return s
}

// getSnapshotAttribute gets the value(s) for the attribute specified with parameter param.
//
// The result is an array of strings with the attribute values.
func getSnapshotAttribute(uuid string, param string) []string {
	// run xe snapshot-list --minimal to get all uuids of vbds of the vm in a comma separated list
	cmd := exec.Command(xeBinary, "snapshot-list", "uuid="+uuid, param, "--minimal")
	// buffer to read cmd output
	var out bytes.Buffer
	//assign cmd stdout to buffer
	cmd.Stdout = &out

	// check command for errors
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	//return array of uuids of all vms
	return strings.Split(string(out.Bytes()), ",")
}

// getVMParents gets all parent uuids of the vm specified by the uuid.
// Result is an array of uuids from all parents of the vm (by uuid).
func getVMParents(uuid string) []string {
	return getVMAttribute(uuid, "params=parent")
}

// getParentDetails gets the value of an attribute of a vm with the specified uuid.
//
// All attributes with its values were returned as map of strings.
func getParentDetails(uuid string) map[string]string {
	var p map[string]string
	p = make(map[string]string)
	p["uuid"] = uuid
	p["selfparent"] = getVMAttribute(uuid, "params=selfparent")[0]
	return p
}

// getVMVbds gets all vbds of the virtual machine with the given uuid.
// Result is an array of uuids of all vbds corresponding to the vm.
func getVMVbds(uuid string) []string {
	// run xe vbd-list --minimal to get all uuids of vbds of the vm in a comma separated list
	cmd := exec.Command(xeBinary, "vbd-list", "vm-uuid="+uuid, "type=Disk", "--minimal")
	// buffer to read cmd output
	var out bytes.Buffer
	//assign cmd stdout to buffer
	cmd.Stdout = &out

	// check command for errors
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	//return array of uuids of all vms
	return strings.Split(string(out.Bytes()), ",")
}

// getVbdAttribute gets the attribute values specified by the parameter param of a vbd specified by the parameter uuid.
func getVbdAttribute(uuid string, param ...string) []string {
	// run xe vbd-list --minimal to get all uuids of vbds of the vm in a comma separated list
	cmd := exec.Command(xeBinary, "vbd-list", "uuid="+uuid, "type=Disk", strings.Join(param, " "), "--minimal")
	// buffer to read cmd output
	var out bytes.Buffer
	//assign cmd stdout to buffer
	cmd.Stdout = &out

	// check command for errors
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	//return array of uuids of all vms
	return strings.Split(string(out.Bytes()), ",")
}

// getVbdDetails gets all details about the vbd with uuid parameter.
func getVbdDetails(uuid string) map[string]string {
	var vb map[string]string
	vb = make(map[string]string)
	vb["uuid"] = uuid
	vb["vdi-name-label"] = getVMAttribute(uuid, "params=")[0]
	return vb
}

// generateExampleXML creates a vm with all supported devices,
// e.g. VBDs, Snapshots, Parents.
// Thsi functions is for testing purposes only.
func generateExampleXML() VirtualMachines {
	Sn := &Snapshots{}
	s1 := &Snapshot{UUID: "980ac4d0-d408-11e9-80e3-84a93e00fae3", NameLable: "Snapshot 1", NameDescription: "", IsVmssSnapshot: "false"}
	s2 := &Snapshot{UUID: "980ac4d0-d408-11e9-80e3-84a93e00fae4", NameLable: "Snapshot 2", NameDescription: "", IsVmssSnapshot: "false"}
	Sn.Snaps = append(Sn.Snaps, *s1)
	Sn.Snaps = append(Sn.Snaps, *s2)

	VBn := &VBDS{}
	vb1 := &VBD{UUID: "31efb927-5130-8be4-1738-dd3051969ea0", VbdType: "Disk", VdiNameLabel: "HDD 0"}
	vb2 := &VBD{UUID: "31efb927-5130-8be4-1738-dd3051969ea1", VbdType: "Disk", VdiNameLabel: "HDD 1"}
	VBn.Vbds = append(VBn.Vbds, *vb1)
	VBn.Vbds = append(VBn.Vbds, *vb2)

	Pn := &Parents{}
	p1 := &Parent{UUID: "e780cf52-3668-1b24-2c6e-7d613b3b502b", Selfparent: false}
	Pn.Ps = append(Pn.Ps, *p1)

	VMn := &VirtualMachines{}
	vm1 := &VirtualMachine{NameLabel: "TestVM", UUID: "fad84dd9-fc06-9cee-7e38-30da28de0cb6", Ps: *Pn, VBDs: *VBn, Snaps: *Sn}
	VMn.Vms = append(VMn.Vms, *vm1)

	return *VMn
}

// generate the XML object for a single vm by its uuid
func generateVMXML(uuid string) VirtualMachine {

	// creating snapshot objects
	Sn := &Snapshots{}
	var tempSnapshots []string = getVMSnapshots(uuid)

	// iterating over all snapshot uuid for vm with uuid
	for index := 0; index < len(tempSnapshots); index++ {
		var tempSnapshot map[string]string = getSnapshotDetails(tempSnapshots[index])
		s := &Snapshot{UUID: tempSnapshot["uuid"], NameLable: tempSnapshot["name-label"], NameDescription: tempSnapshot["name-description"], IsVmssSnapshot: tempSnapshot["is-vmss-snapshot"]}
		Sn.Snaps = append(Sn.Snaps, *s)
	}

	// creating VBDs objects
	VBn := &VBDS{}
	var tempVbds []string = getVMVbds(uuid)

	// iterating over all vbd uuid for vm with uuid
	for index := 0; index < len(tempSnapshots); index++ {
		var tempVbd map[string]string = getVbdDetails(tempVbds[index])
		vb := &VBD{UUID: tempVbd["uuid"], VbdType: tempVbd["type"], VdiNameLabel: tempVbd["vdi-name-label"]}
		VBn.Vbds = append(VBn.Vbds, *vb)
	}

	// creating Parent objects
	Pn := &Parents{}
	var tempParents []string = getVMParents(uuid)

	// iterating over all parent uuid for parent with uuid
	for index := 0; index < len(tempParents); index++ {
		var tempParent map[string]string = getParentDetails(tempParents[index])
		ptemp, _ := strconv.ParseBool(tempParent["selfparent"])
		p := &Parent{UUID: tempParent["uuid"], Selfparent: ptemp}
		Pn.Ps = append(Pn.Ps, *p)
	}

	// creating our vm with all attributes
	var tempVM map[string]string = getVMDetails(uuid)
	vm := &VirtualMachine{NameLabel: tempVM["name-label"], UUID: tempVM["uuid"], Ps: *Pn, VBDs: *VBn, Snaps: *Sn}

	return *vm
}

// generateVmsXML creates the object of virtual machines for XML generation.
// As parameter is an array of uuids of the virtual machines given.
func generateVmsXML(uuids []string) VirtualMachines {
	VMn := &VirtualMachines{}
	// we iterate over all uuids
	for index := 0; index < len(uuids); index++ {
		// create an vm object
		var tempVM VirtualMachine = generateVMXML(uuids[index])
		// appended the created vm object to the result object
		VMn.Vms = append(VMn.Vms, tempVM)
	}
	return *VMn
}

func main() {
	// ################## variable definitions

	// parse command line options
	parseCommandOptions()

	// to generate an example XML file with one VM comment out the following line
	vms := generateVmsXML(getVMs("power-state=running"))

	// create the export file
	file, _ := os.Create(exportFile)

	// get an writer for the export file
	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")

	if err := enc.Encode(vms); err != nil {
		fmt.Printf("error: %vms\n", err)
	}
	//fmt.Printf("--> VMs %+v", vms)
}
