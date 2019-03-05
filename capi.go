package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os/exec"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"io/ioutil"
	"fmt"
	"bytes"
	. "github.com/mattn/go-getopt"
	"reflect"
)

var lock = sync.RWMutex{}
var listen string

// The bcreate Type. Name of elements must match with jconf params
type Bcreate struct {
	Jname			string	`json:"jname,omitempty"`
	Xhci			string	`json:"xhci,omitempty"`
	Astart			string	`json:"astart,omitempty"`
	Relative_path		string	`json:"relative_path,omitempty"`
	Path			string	`json:"path,omitempty"`
	Data			string	`json:"data,omitempty"`
	Rcconf			string	`json:"rcconf,omitempty"`
	Host_hostname		string	`json:"host_hostname,omitempty"`
	Ip4_addr		string	`json:"ip4_addr,omitempty"`
	Nic_hwaddr		string	`json:"nic_hwaddr,omitempty"`
	Zfs_snapsrc		string	`json:"zfs_snapsrc,omitempty"`
	Runasap			string	`json:"runasap,omitempty"`
	Interface		string	`json:"interface,omitempty"`
	Rctl_nice		string	`json:"rctl_nice,omitempty"`
	Emulator		string	`json:"emulator,omitempty"`
	Imgsize			string	`json:"imgsize,omitempty"`
	Imgtype			string	`json:"imgtype,omitempty"`
	Vm_cpus			string	`json:"vm_cpus,omitempty"`
	Vm_ram			string	`json:"vm_ram,omitempty"`
	Vm_os_type		string	`json:"vm_os_type,omitempty"`
	Vm_efi			string	`json:"vm_efi,omitempty"`
	Iso_site		string	`json:"iso_site,omitempty"`
	Iso_img			string	`json:"iso_img,omitempty"`
	Register_iso_name	string	`json:"register_iso_name,omitempty"`
	Register_iso_as		string	`json:"register_iso_as,omitempty"`
	Vm_hostbridge		string	`json:"vm_hostbridge,omitempty"`
	Bhyve_flags		string	`json:"bhyve_flags,omitempty"`
	Virtio_type		string	`json:"virtio_type,omitempty"`
	Vm_os_profile		string	`json:"vm_os_profile,omitempty"`
	Swapsize		string	`json:"swapsize,omitempty"`
	Vm_iso_path		string	`json:"vm_iso_path,omitempty"`
	Vm_guestfs		string	`json:"vm_guestfs,omitempty"`
	Vm_vnc_port		string	`json:"vm_vnc_port,omitempty"`
	Bhyve_generate_acpi	string	`json:"bhyve_generate_acpi,omitempty"`
	Bhyve_wire_memory	string	`json:"bhyve_wire_memory,omitempty"`
	Bhyve_rts_keeps_utc	string	`json:"bhyve_rts_keeps_utc,omitempty"`
	Bhyve_force_msi_irq	string	`json:"bhyve_force_msi_irq,omitempty"`
	Bhyve_x2apic_mode	string	`json:"bhyve_x2apic_mode,omitempty"`
	Bhyve_mptable_gen	string	`json:"bhyve_mptable_gen,omitempty"`
	Bhyve_ignore_msr_acc	string	`json:"bhyve_ignore_msr_acc,omitempty"`
	Cd_vnc_wait		string	`json:"cd_vnc_wait,omitempty"`
	Bhyve_vnc_resolution	string	`json:"bhyve_vnc_resolution,omitempty"`
	Bhyve_vnc_tcp_bind	string	`json:"bhyve_vnc_tcp_bind,omitempty"`
	Bhyve_vnc_vgaconf	string	`json:"bhyve_vnc_vgaconf,omitempty"`
	Nic_driver		string	`json:"nic_driver,omitempty"`
	Vnc_password		string	`json:"vnc_password,omitempty"`
	Media_auto_eject	string	`json:"media_auto_eject,omitempty"`
	Vm_cpu_topology		string	`json:"vm_cpu_topology,omitempty"`
	Debug_engine		string	`json:"debug_engine,omitempty"`
	Cd_boot_firmware	string	`json:"cd_boot_firmware,omitempty"`
	Jailed			string	`json:"jailed,omitempty"`
	On_poweroff		string	`json:"on_poweroff,omitempty"`
	On_reboot		string	`json:"on_reboot,omitempty"`
	On_crash		string	`json:"on_crash,omitempty"`
}

var people []Bcreate

type Bhyves struct {
	Jname     string
	Jid       int
	Vm_Ram    int // MB
	Vm_Cpus   int
	Vm_Os_Type string
	Status string
	Vnc string
}

var bhyves []Bhyves

func init() {
	var c int
	// defaults
	listen = ":8080"

	OptErr = 0
	for {
		if c = Getopt("l:h"); c == EOF {
			break
		}
		switch c {
		case 'l':
			listen = OptArg
		case 'h':
			usage()
			os.Exit(1)
		}
	}
}

func usage() {
	println("usage: capi [-l listenaddress|-h]")
}

// main function to boot up everything
func main() {

	HandleInitBhyveList()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blist", HandleBhyveList).Methods("GET")
	router.HandleFunc("/api/v1/cacheblist", HandleCacheBhyveList).Methods("GET")
	router.HandleFunc("/api/v1/bstart/{instanceid}", HandleBhyveStart).Methods("POST")
	router.HandleFunc("/api/v1/bstop/{instanceid}", HandleBhyveStop).Methods("POST")
	router.HandleFunc("/api/v1/bremove/{instanceid}", HandleBhyveRemove).Methods("POST")
	router.HandleFunc("/api/v1/bcreate/{instanceid}", HandleBhyveCreate).Methods("POST")
	log.Fatal(http.ListenAndServe(listen, router))
}

func HandleBhyveList(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	cmd := exec.Command("env","NOCOLOR=0","cbsd","bls","header=0","display=jname,jid,vm_ram,vm_cpus,vm_os_type,status,vnc_port")
	stdout, err := cmd.Output()
	lock.Unlock()

	if err != nil {
		return
	}

	lines := strings.Split(string(stdout), "\n")
	imas := make([]Bhyves, 0)

	for _, line := range lines {
		if len(line)>2 {
			ima := Bhyves{}
			re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
			re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			n := re_leadclose_whtsp.ReplaceAllString(line, "")
			n = re_inside_whtsp.ReplaceAllString(line, " ")
			ima.Jname = strings.Split(n, " ")[0]
			ima.Jid, err = strconv.Atoi(strings.Split(n, " ")[1])
			ima.Vm_Ram, err = strconv.Atoi(strings.Split(n, " ")[2])
			ima.Vm_Cpus, err = strconv.Atoi(strings.Split(n, " ")[3])
			ima.Vm_Os_Type = strings.Split(n, " ")[4]
			ima.Status = strings.Split(n, " ")[5]
			ima.Vnc = strings.Split(n, " ")[6]
			imas = append(imas, ima)
		}
	}

	//w.WriteJson(imas)
	json.NewEncoder(w).Encode(&imas)
}

func HandleInitBhyveList() {
	lock.Lock()
	cmd := exec.Command("env","NOCOLOR=0","cbsd","bls","header=0","display=jname,jid,vm_ram,vm_cpus,vm_os_type,status,vnc_port")
	stdout, err := cmd.Output()
	lock.Unlock()

	if err != nil {
		return
	}

	lines := strings.Split(string(stdout), "\n")

	for _, line := range lines {
		if len(line)>2 {
			ima := Bhyves{}
			re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
			re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			n := re_leadclose_whtsp.ReplaceAllString(line, "")
			n = re_inside_whtsp.ReplaceAllString(line, " ")
			ima.Jname = strings.Split(n, " ")[0]
			ima.Jid, err = strconv.Atoi(strings.Split(n, " ")[1])
			ima.Vm_Ram, err = strconv.Atoi(strings.Split(n, " ")[2])
			ima.Vm_Cpus, err = strconv.Atoi(strings.Split(n, " ")[3])
			ima.Vm_Os_Type = strings.Split(n, " ")[4]
			ima.Status = strings.Split(n, " ")[5]
			ima.Vnc = strings.Split(n, " ")[6]
			bhyves = append(bhyves, ima)
		}
	}
}

func HandleCacheBhyveList(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(&bhyves)
}

func HandleBhyveStart(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var instanceid string
	_ = json.NewDecoder(r.Body).Decode(&instanceid)
	instanceid = params["instanceid"]

	go realInstanceStart(instanceid)

	return
}

func realInstanceStart(instanceid string) {
	jname := "jname=" + instanceid
	cmd := exec.Command("env","NOCOLOR=0","cbsd","bstart","inter=0", jname)
	_, err := cmd.Output()

	if err != nil {
		return
	}
}


func HandleBhyveStop(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var instanceid string
	buf, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		fmt.Printf("bodyErr %s", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	fmt.Printf("BODY rdr1: %q", rdr1)
	r.Body = rdr2

	_ = json.NewDecoder(r.Body).Decode(&instanceid)
	instanceid = params["instanceid"]

	go realInstanceStop(instanceid)

	return
}

func realInstanceStop(instanceid string) {
	jname := "jname=" + instanceid
	cmd := exec.Command("env","NOCOLOR=0","cbsd","bstop","inter=0", jname)
	_, err := cmd.Output()

	if err != nil {
		return
	}
}


func HandleBhyveRemove(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var instanceid string
	buf, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		//log.Print("bodyErr ", bodyErr.Error())
		fmt.Printf("bodyErr %s", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	fmt.Printf("BODY rdr1: %q", rdr1)
	//fmt.Printf("BODY rdr2: %q", rdr2)
	r.Body = rdr2

	_ = json.NewDecoder(r.Body).Decode(&instanceid)
	instanceid = params["instanceid"]
	go realInstanceRemove(instanceid)

	return
}

func realInstanceRemove(instanceid string) {
	jname := "jname=" + instanceid
	cmd := exec.Command("env","NOCOLOR=0","cbsd","bremove","inter=0", jname)
	_, err := cmd.Output()

	if err != nil {
		return
	}
}


func realInstanceCreate(createstr string) {

	fmt.Printf("bcreate: [ %s ]",createstr);

	createstr = strings.TrimSuffix(createstr, "\n")
	arrCommandStr := strings.Fields(createstr)
	cmd := exec.Command(arrCommandStr[0], arrCommandStr[1:]...)

	cmd.Stdin = os.Stdin;
	cmd.Stdout = os.Stdout;
	cmd.Stderr = os.Stderr;

	err := cmd.Run() 

	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func getStructTag(f reflect.StructField) string {
	return string(f.Tag)
}

func HandleBhyveCreate(w http.ResponseWriter, r *http.Request) {
	var instanceid string
	params := mux.Vars(r)
	instanceid = params["instanceid"]

	if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
	}

	var bcreate Bcreate
	_ = json.NewDecoder(r.Body).Decode(&bcreate)
	bcreate.Jname = instanceid
	json.NewEncoder(w).Encode(bcreate)
	val := reflect.ValueOf(bcreate)

//	fmt.Println("J ",bcreate.Jname)
//	fmt.Println("R ",bcreate.Vm_ram)

	var jconf_param string
	var str strings.Builder

	str.WriteString("env NOCOLOR=1 /usr/local/bin/cbsd bcreate inter=0 ")

	for i := 0; i < val.NumField(); i++ {
		//fmt.Printf("TEST %d\n",i);
		valueField := val.Field(i)

		typeField := val.Type().Field(i)
		tag := typeField.Tag

		tmpval := fmt.Sprintf("%s",valueField.Interface())

		if len(tmpval) == 0 {
			continue
		}

		jconf_param = strings.ToLower(typeField.Name)
		fmt.Printf("jconf: %s,\tField Name: %s,\t Field Value: %v,\t Tag Value: %s\n", jconf_param, typeField.Name, valueField.Interface(), tag.Get("tag_name"))
		buf := fmt.Sprintf("%s=%v ", jconf_param, tmpval)
		str.WriteString(buf)
	}

	go realInstanceCreate(str.String())

	return
}
