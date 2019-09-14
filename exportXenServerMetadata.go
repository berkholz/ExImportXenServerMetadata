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

type VirtualMachines struct {
	XMLName xml.Name         `xml:"vms"`
	Vms     []VirtualMachine `xml:"vm"`
}
type VirtualMachine struct {
	XMLName   xml.Name  `xml:"vm"`
	NameLabel string    `xml:"name,attr"`
	UUID      string    `xml:"uuid,attr"`
	Ps        Parents   `xml:"parents"`
	VBDs      VBDS      `xml:"vbds"`
	Snaps     Snapshots `xml:"snapshots"`
}

// list of parents of a vm
type Parents struct {
	XMLName xml.Name `xml:"parents"`
	Ps      []Parent `xml:"parent"`
}

// parent of a vm, part of the list
type Parent struct {
	XMLName    xml.Name `xml:"parent"`
	UUID       string   `xml:"uuid,attr"`
	Selfparent bool     `xml:"selfparent,attr"`
}

// list of vbds of a vm
type VBDS struct {
	XMLName xml.Name `xml:"vbds"`
	Vbds    []VBD    `xml:"vbd"`
}

// vbd of a vm, part of the list
type VBD struct {
	XMLName      xml.Name `xml:"vbd"`
	UUID         string   `xml:"uuid,attr"`
	VbdType      string   `xml:"type,attr"`
	VdiNameLabel string   `xml:"vdi-name-label,attr"`
}

// list of snapshots of a vm
type Snapshots struct {
	XMLName xml.Name   `xml:"snapshots"`
	Snaps   []Snapshot `xml:"snapshot"`
}

// snapshot of a vm, part of the list
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

// getVms returns an array of all uuid of vms via xe command line.
// The set of vms can be filtered by xe command line filter, e.g. power-state=running.
// For filter, please consult the xe command line reference:
//  https://docs.citrix.com/en-us/citrix-hypervisor/command-line-interface.html
func getVms(xefilter string) []string {
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
// * power-state=running
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

func getVmData(uuid string, param string) string {

	// run xe vm-list --minimal to get all uuids of vms in a comma separated list
	cmd := exec.Command(xeBinary, "vm-list", "uuid="+uuid, "params="+param, "--minimal")
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
	return string(out.Bytes())
}

func getVm(uuid string) map[string]string {
	var vm map[string]string
	vm = make(map[string]string)
	vm["uuid"] = uuid
	vm["name-label"] = getVmData(uuid, "name-label")
	return vm
}

func getVmSnapshots(uuid string) []string {
	return nil
}

func getVmSnapshot(uuid string) map[string]string {
	return nil
}

func getVmVbds(uuid string) []string {
	return nil
}

func getVmVbd(uuid string) map[string]string {
	return nil
}

func getVmParents(uuid string) []string {
	return nil
}

func getVmParent(uuid string) map[string]string {
	return nil
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
func generateVmXML(uuid string) VirtualMachine {

	// creating snapshot objects
	Sn := &Snapshots{}
	var tempSnapshots []string = getVmSnapshots(uuid)

	// iterating over all snapshot uuid for vm with uuid
	for index := 0; index < len(tempSnapshots); index++ {
		var tempSnapshot map[string]string = getVmSnapshot(tempSnapshots[index])
		s := &Snapshot{UUID: tempSnapshot["uuid"], NameLable: tempSnapshot["name-label"], NameDescription: tempSnapshot["name-description"], IsVmssSnapshot: tempSnapshot["is-vmss-snapshot"]}
		Sn.Snaps = append(Sn.Snaps, *s)
	}

	// creating VBDs objects
	VBn := &VBDS{}
	var tempVbds []string = getVmVbds(uuid)

	// iterating over all vbd uuid for vm with uuid
	for index := 0; index < len(tempSnapshots); index++ {
		var tempVbd map[string]string = getVmVbd(tempVbds[index])
		vb := &VBD{UUID: tempVbd["uuid"], VbdType: tempVbd["type"], VdiNameLabel: tempVbd["vdi-name-label"]}
		VBn.Vbds = append(VBn.Vbds, *vb)
	}

	// creating Parent objects
	Pn := &Parents{}
	var tempParents []string = getVmParents(uuid)

	// iterating over all parent uuid for parent with uuid
	for index := 0; index < len(tempParents); index++ {
		var tempParent map[string]string = getVmParent(tempParents[index])
		ptemp, _ := strconv.ParseBool(tempParent["selfparent"])
		p := &Parent{UUID: tempParent["uuid"], Selfparent: ptemp}
		Pn.Ps = append(Pn.Ps, *p)
	}

	// creating our vm with all attributes
	var tempVm map[string]string = getVm(uuid)
	vm := &VirtualMachine{NameLabel: tempVm["name-label"], UUID: tempVm["uuid"], Ps: *Pn, VBDs: *VBn, Snaps: *Sn}

	return *vm
}

func generateVmsXML(uuids []string) VirtualMachines {
	VMn := &VirtualMachines{}
	for index := 0; index < len(uuids); index++ {
		var tempVm VirtualMachine = generateVmXML(uuids[index])
		VMn.Vms = append(VMn.Vms, tempVm)
	}

	return *VMn
}

func main() {
	// ################## variable definitions

	// parse command line options
	parseCommandOptions()

	// to generate an example XML file with one VM comment out the following line
	vms := generateExampleXML()

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
