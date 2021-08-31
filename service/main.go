package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	v1b1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// EMDC represents k8s cluster
type EMDC struct {
	ID       int    `json:"id"`
	Host     string `json:"host"`
	confFile string
}

// ServiceEntry from istio
type ServiceEntry struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	PortNumber uint32 `json:"portnumber"`
	Protocol   string `json:"protocol"`
	Host       string `json:"host"`
	Namespace  string `json:"namespace"`
}

var currentID int
var emdcs map[int]EMDC

func repoInit() {
	currentID = 0
	emdcs = make(map[int]EMDC)
}

func repoCreateEMDC(host string, confFile string) EMDC {
	// TODO check if EMDC for this host already exists
	emdc := EMDC{ID: currentID, Host: host, confFile: confFile}
	emdcs[currentID] = emdc
	currentID++
	return emdc
}

func repoReturnAllEMDCs() []EMDC {
	ret := make([]EMDC, 0, len(emdcs))
	for _, emdc := range emdcs {
		ret = append(ret, emdc)
	}
	return ret
}

func repoGetEMDC(id int) *EMDC {
	if val, ok := emdcs[id]; ok {
		return &val
	}
	return nil
}

func repoDeleteEMDC(id int) {
	delete(emdcs, id)
}

func returnAllEMDCs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repoReturnAllEMDCs())
}

func createNewEMDC(w http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("emdc")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("/tmp", "upload-*.yml")
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	config, err := clientcmd.BuildConfigFromFlags("", tempFile.Name())
	if err != nil {
		// TODO delete temp file
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	emdc := repoCreateEMDC(config.Host, tempFile.Name())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(emdc)
}

func returnSingleEMDC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if emdc := repoGetEMDC(id); emdc != nil {
		json.NewEncoder(w).Encode(repoGetEMDC(id))
		return
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func deleteEMDC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	repoDeleteEMDC(id)
}

func returnAllServiceEntries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	emdc := repoGetEMDC(id)
	if emdc == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	config, err := clientcmd.BuildConfigFromFlags("", emdc.confFile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ic, err := versionedclient.NewForConfig(config)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	seList, err := ic.NetworkingV1beta1().ServiceEntries("").List(context.TODO(), metav1.ListOptions{})
	ret := make([]ServiceEntry, 0, len(seList.Items))
	for _, se := range seList.Items {
		ret = append(ret, ServiceEntry{
			UUID:       string(se.UID),
			Host:       se.Spec.Hosts[0],
			Namespace:  se.Namespace,
			Name:       se.Name,
			PortNumber: se.Spec.Ports[0].Number,
			Protocol:   se.Spec.Ports[0].Protocol,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}

// curl -H "Content-Type: application/json" -d '{"host":"plot.gpu.global", "name":"fira", "namespace":"x86", "portnumber":80, "protocol":"http"}' http://localhost:8090/emdc/0/serviceentry/create/1
func createNewServiceEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	id1, err := strconv.Atoi(vars["id1"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id2, err := strconv.Atoi(vars["id2"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	emdc1 := repoGetEMDC(id1)
	if emdc1 == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	emdc2 := repoGetEMDC(id2)
	if emdc2 == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	config1, err := clientcmd.BuildConfigFromFlags("", emdc1.confFile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ic, err := versionedclient.NewForConfig(config1)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	config2, err := clientcmd.BuildConfigFromFlags("", emdc2.confFile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	clientset2, err := kubernetes.NewForConfig(config2)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	namespace := "istio-system"
	pods, err := clientset2.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "istio=ingressgateway"})
	ingressIP := pods.Items[0].Status.HostIP
	var ingressPort int32
	s, err := clientset2.CoreV1().Services(namespace).Get(context.TODO(), "istio-ingressgateway", metav1.GetOptions{})
	for _, p := range s.Spec.Ports {
		if p.Port == 15443 {
			ingressPort = p.NodePort
			break
		}
	}
	var svcEntry ServiceEntry
	if err := json.Unmarshal(body, &svcEntry); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	var se v1beta1.ServiceEntry
	se.ObjectMeta.Name = svcEntry.Name
	se.Spec.Hosts = []string{svcEntry.Host}
	se.Spec.Location = 1   // MESH_INTERNAL
	se.Spec.Resolution = 2 // DNS
	portName := "braine"
	se.Spec.Ports = []*v1b1.Port{{Number: svcEntry.PortNumber, Protocol: svcEntry.Protocol, Name: portName}}

	ip := fmt.Sprintf("240.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256))
	se.Spec.Addresses = []string{ip}
	we := v1b1.WorkloadEntry{Address: ingressIP, Ports: map[string]uint32{portName: uint32(ingressPort)}}
	se.Spec.Endpoints = []*v1b1.WorkloadEntry{&we}
	ret, err := ic.NetworkingV1beta1().ServiceEntries(svcEntry.Namespace).Create(context.TODO(), &se, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	resp := ServiceEntry{
		UUID:       string(ret.UID),
		Host:       ret.Spec.Hosts[0],
		Namespace:  ret.Namespace,
		Name:       ret.Name,
		PortNumber: ret.Spec.Ports[0].Number,
		Protocol:   ret.Spec.Ports[0].Protocol,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func returnSingleServiceEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	uuid := vars["uuid"]
	emdc := repoGetEMDC(id)
	if emdc == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	config, err := clientcmd.BuildConfigFromFlags("", emdc.confFile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ic, err := versionedclient.NewForConfig(config)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	seList, err := ic.NetworkingV1beta1().ServiceEntries("").List(context.TODO(), metav1.ListOptions{})
	for _, se := range seList.Items {
		if string(se.UID) == uuid {
			ret := ServiceEntry{
				UUID:       uuid,
				Host:       se.Spec.Hosts[0],
				Namespace:  se.Namespace,
				Name:       se.Name,
				PortNumber: se.Spec.Ports[0].Number,
				Protocol:   se.Spec.Ports[0].Protocol,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ret)
			return
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func deleteServiceEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	uuid := vars["uuid"]
	emdc := repoGetEMDC(id)
	if emdc == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	config, err := clientcmd.BuildConfigFromFlags("", emdc.confFile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ic, err := versionedclient.NewForConfig(config)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	seList, err := ic.NetworkingV1beta1().ServiceEntries("").List(context.TODO(), metav1.ListOptions{})
	for _, se := range seList.Items {
		if string(se.UID) == uuid {
			err = ic.NetworkingV1beta1().ServiceEntries(se.Namespace).Delete(context.TODO(), se.Name, metav1.DeleteOptions{})
			return
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func main() {
	repoInit()
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/emdcs", returnAllEMDCs)
	myRouter.HandleFunc("/emdc", createNewEMDC).Methods("POST")
	myRouter.HandleFunc("/emdc/{id}", deleteEMDC).Methods("DELETE")
	myRouter.HandleFunc("/emdc/{id}", returnSingleEMDC)

	myRouter.HandleFunc("/emdc/{id}/serviceentries", returnAllServiceEntries)
	myRouter.HandleFunc("/emdc/{id1}/serviceentry/create/{id2}", createNewServiceEntry).Methods("POST")
	myRouter.HandleFunc("/emdc/{id}/serviceentry/{uuid}", deleteServiceEntry).Methods("DELETE")
	myRouter.HandleFunc("/emdc/{id}/serviceentry/{uuid}", returnSingleServiceEntry)
	log.Fatal(http.ListenAndServe(":8090", myRouter))
}
