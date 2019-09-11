package main

// ################## imports
import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
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

	if len(exportFile) == 0 {
		exportFile = "vms.export.example.xml"
		fmt.Println("No export file given, using default: " + exportFile)
	}
}

func generateExampleXML() VirtualMachines {
	Sn := &Snapshots{}
	s1 := &Snapshot{UUID: "980ac4d0-d408-11e9-80e3-84a93e00fae3", NameLable: "ADM-WIKI", NameDescription: "", IsVmssSnapshot: "false"}
	s2 := &Snapshot{UUID: "980ac4d0-d408-11e9-80e3-84a93e00fae4", NameLable: "ADM-WIKI2", NameDescription: "", IsVmssSnapshot: "false"}
	Sn.Snaps = append(Sn.Snaps, *s1)
	Sn.Snaps = append(Sn.Snaps, *s2)

	VBn := &VBDS{}
	vb1 := &VBD{UUID: "31efb927-5130-8be4-1738-dd3051969ea0", VbdType: "Disk", VdiNameLabel: "ADM-WIKI 0"}
	vb2 := &VBD{UUID: "31efb927-5130-8be4-1738-dd3051969ea1", VbdType: "Disk", VdiNameLabel: "ADM-WIKI 1"}
	VBn.Vbds = append(VBn.Vbds, *vb1)
	VBn.Vbds = append(VBn.Vbds, *vb2)

	Pn := &Parents{}
	p1 := &Parent{UUID: "e780cf52-3668-1b24-2c6e-7d613b3b502b", Selfparent: false}
	Pn.Ps = append(Pn.Ps, *p1)

	VMn := &VirtualMachines{}
	vm1 := &VirtualMachine{NameLabel: "ADM-XWiki", UUID: "fad84dd9-fc06-9cee-7e38-30da28de0cb6", Ps: *Pn, VBDs: *VBn, Snaps: *Sn}
	VMn.Vms = append(VMn.Vms, *vm1)

	return *VMn
}

func main() {
	// ################## variable definitions

	// parse command line options
	parseCommandOptions()

	// we initialize our Users array
	vms := generateExampleXML()

	file, _ := os.Create(exportFile)

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")

	if err := enc.Encode(vms); err != nil {
		fmt.Printf("error: %vms\n", err)
	}
	//fmt.Printf("--> VMs %+v", vms)
}
