package waitingloop

import (
	// "bytes"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	// "path"
	// "path/filepath"
	// "sort"
	"strings"
	// "sigs.k8s.io/kustomize/kyaml/kio"
	// "sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	// "sigs.k8s.io/kustomize/kyaml/yaml"
	"net/http"
)

const providerApiServiceUrl string = "http://provider-api-svc:3333"
const kubeAdmControlPlaneEndpoint string = "/getkubeadmcontrolplanes"
const kubeConfigEndpoint string = "/getkubeconfig"

type Message struct {
	Namespace string `json:"Namespace,omitempty"`
	Name      string `json:"Name,omitempty"`
	Phase     string `json:"Phase,omitempty"`
	Age       string `json:"Age,omitempty"`
}
type KubeConfig struct {
	Name       string
	Namespace  string
	Kubeconfig string
}
type KubeConfigList []KubeConfig

func getStatus(name string) (string, bool) {
	fmt.Println("Get status of cluster %s", name)
	request, error := http.NewRequest("GET", providerApiServiceUrl+kubeAdmControlPlaneEndpoint, bytes.NewBuffer([]byte("abc")))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		return "error occurred", true
	}
	fmt.Println("Response Status:", response.Status)
	httpPostBody, err := ioutil.ReadAll(response.Body) //<--- here!

	if err != nil {
		fmt.Println("Read response error: %s", err)
		return string(err.Error()), true
	}
	if len(httpPostBody) < 5 {
		return "error occurred", true
	}

	return string(httpPostBody), false

}
func getKubeConfig(namespace string, name string) (string, bool) {
	fmt.Println("Get kubeconfig of cluster %s", name)
	request, error := http.NewRequest("GET", providerApiServiceUrl+kubeConfigEndpoint, bytes.NewBuffer([]byte("abc")))
	request.Header.Set("clustername", name)
	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		return "error occurred", true
	}
	fmt.Println("Response Status:", response.Status)
	httpPostBody, err := ioutil.ReadAll(response.Body) //<--- here!

	if err != nil {
		fmt.Println("Read response error: %s", err)
		return string(err.Error()), true
	}
	if len(httpPostBody) < 5 {
		return "error occurred", true
	}

	return string(httpPostBody), false
}
func RunWaitingLoop(namespace string, name string, communicate chan string) (string, bool) {
	fmt.Println("Waiting Loop for Cluster Creation Process")
	fmt.Println("Cluster Name: %s \n", strings.ToUpper(name))
	// gocron.Every(5).Seconds().Do(task)
	// Get status of cluster
	provisionedCluster := false
	for !provisionedCluster {
		jsonStringKubeAdmRespone, err := getStatus(name)
		var jsonKubeAdmResponse []Message
		if !err {
			error := json.Unmarshal([]byte(jsonStringKubeAdmRespone), &jsonKubeAdmResponse)
			if error != nil {
				fmt.Println("COnvertting json got error: ", error)
			}
		} else {
			fmt.Println("Error get Status cluster: %s", err)
			communicate <- "error"
			return "error", true
		}

		// Parse status of cluster by cluster name

		// If cluster is provisioned => get kubeconfig
		// If not => sleep 5s
		for _, item := range jsonKubeAdmResponse {
			if item.Namespace == namespace && item.Name == name {
				// Check status of cluster:
				if item.Phase == "Provisioned" {
					fmt.Println("Cluster %s is provisioned.. get Kubeconfig", name)
					// get Kubeconfig
					kubeconfig, err := getKubeConfig(namespace, name)
					if err {
						fmt.Println("Get KubeConfig cluster name %s namespace %s failed", name, namespace)
						communicate <- kubeconfig
						return kubeconfig, true
					}
					communicate <- "error"
					return kubeconfig, false
					// return kubeconfig
				}
			}
		}
		// Keep waiting for cluster change to status provisioned
		time.Sleep(5 * time.Second)
	}
	return "error", true
}
