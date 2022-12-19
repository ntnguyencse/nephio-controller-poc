package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	// work "github.com/gocraft/work"
	// "container/list"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	yamlFileTemplate "github.com/ntnguyencse/nephio-controller-poc/provider-api-service/yaml-template"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v2"
)

// Define listening port
const serverPort string = ":3333"
const kubectlCmd string = "kubectl"
const clusterctlCmd string = "clusterctl"

var kubeConfig string
var namespaceClusterAPI string

// Structure for Parse cloud yaml file
type CloudYaml struct {
	Clouds map[string]CloudInformation `yaml:"clouds"`
}
type AuthStruct struct {
	AuthUrl           string `yaml:"auth_url"`
	ProjectName       string `yaml:"project_name"`
	UserName          string `yaml:"username"`
	Password          string `yaml:"password"`
	UserDomainName    string `yaml:"user_domain_name"`
	ProjectDomainName string `yaml:"project_domain_name"`
}
type CloudInformation struct {
	RegionName string     `yaml:"region_name"`
	AuthInform AuthStruct `yaml:"auth"`
}

// End of parse yaml file structure
type KubeConfigMessage struct {
	Name       string `json:"Name"`
	KubeConfig string `json:"KubeConfig"`
}
type Message struct {
	Namespace string `json:"Namespace,omitempty"`
	Name      string `json:"Name,omitempty"`
	Phase     string `json:"Phase,omitempty"`
	Age       string `json:"Age,omitempty"`
}
type ClusterConfigurations struct {
	ClusterName              string `json:"ClusterName"`
	KubernetesVersion        string `json:"KubernetesVersion"`
	ControlPlaneMachineCount string `json:"ControlPlaneMachineCount"`
	KubernetesMachineCount   string `json:"KubernetesMachineCount"`
}
type ClusterRecord struct {
	Name                     string            `json:"name,omitempty"`
	InfraType                string            `json:"infraType,omitempty"`
	Labels                   map[string]string `json:"labels,omitempty"`
	Repository               string            `json:"repository,omitempty"`
	Provider                 string            `json:"provider,omitempty"`
	ProvisionMethod          string            `json:"provisionMethod,omitempty"`
	Namespace                string            `json:"namespace,omitempty"`
	KubernetesVersion        string            `json:"pubernetesVersion,omitempty"`
	ControlPlaneMachineCount string            `json:"controlPlaneMachineCount,omitempty"`
	KubernetesMachineCount   string            `json:"kubernetesMachineCount,omitempty"`
	CreatedTime              time.Time         `json:"createdTime,omitempty"`
	UpdatedTime              time.Time         `json:"updatedTime,omitempty"`
}
type Machine struct {
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
	NodeName   string `json:"nodename,omitempty"`
	ProviderID string `json:"providerid,omitempty"`
	Phase      string `json:"phase,omitempty"`
	Age        string `json:"age,omitempty"`
	Version    string `json:"version,omitempty"`
}

type Machines struct {
	List []Machine `json:"machines,omitempty"`
}

var listYamlFileClusterAPI []string

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewSHA1Hash generates a new SHA1 hash based on
// a random number of characters.
func NewSHA1Hash(n ...int) string {
	noRandomCharacters := 32

	if len(n) > 0 {
		noRandomCharacters = n[0]
	}

	randString := RandomString(noRandomCharacters)

	hash := sha1.New()
	hash.Write([]byte(randString))
	bs := hash.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

var characterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// RandomString generates a random string of n length
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characterRunes[rand.Intn(len(characterRunes))]
	}
	return string(b)
}

func main() {
	// kubectlExecutablePath, _ := exec.LookPath("kubectl")

	// currentListCluster := list.newList()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// fmt.Println("KubeConfig file path" + os.Getenv("KUBECONFIG"))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		// construct `go version` command
		// cmdGoVer := &exec.Cmd{
		// 	Path:   kubectlExecutablePath,
		// 	Args:   []string{kubectlExecutablePath, "version"},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stdout,
		// }

		// // run `go version` command
		// if err := cmdGoVer.Run(); err != nil {
		// 	fmt.Println("Error:", err)
		// }
		// command := Command("kubectl", "version")
		// // set var to get the output
		// var out bytes.Buffer

		// // set the output to our variable
		// command.Stdout = &out
		// err := command.Run()
		// if err != nil {
		// 	log.Println(err)
		// }

		// fmt.Println(out.String())
		res, err := exec.Command("./kubectl", "version").Output()
		if err != nil {
			panic(err)
		}
		fmt.Printf("OUTPUT: %s", res)
		resclusterapi, errclapi := exec.Command("./clusterctl", "version").Output()
		if errclapi != nil {
			panic(errclapi)
		}
		fmt.Printf("OUTPUT: %s", resclusterapi)
		w.Write([]byte(string("clusterctl version\n:" + string(resclusterapi) + string("\n kubectl version: \n") + string(res))))
	})
	r.Get("/getcluster", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received Get Cluster Request")
		prg := "./kubectl"
		arg1 := "get"
		arg2 := "cluster"
		arg3 := "-A"
		argKubeConfig := "--kubeconfig"
		cmd := exec.Command(prg, arg1, arg2, arg3, argKubeConfig, kubeConfig)
		stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			// log.Fatal(err)
			return
		}

		var getClusterResult []Message
		trimmedString := strings.TrimSpace(string(stdout))
		listTrimmedString := strings.Split(trimmedString, "\n")
		if len(listTrimmedString) < 2 {
			w.Write([]byte(string(stdout)))
		}
		for i, str := range listTrimmedString {
			if i != 0 {
				splitStr := strings.Fields(str)
				msg := Message{splitStr[0], splitStr[1], splitStr[2], splitStr[3]}
				msgMarshaled, _ := json.Marshal(msg)
				fmt.Println("msgMarshaled", string(msgMarshaled))
				getClusterResult = append(getClusterResult, msg)
			}
		}
		jsongetClusterResult, errorConvertJson := json.Marshal(getClusterResult)
		if errorConvertJson != nil {
			fmt.Println("error:", errorConvertJson)
		}

		w.Write([]byte(string(jsongetClusterResult)))
	})

	r.Get("/getkubeadmcontrolplanes", func(w http.ResponseWriter, r *http.Request) {

		prg := "./kubectl"
		arg1 := "get"
		arg2 := "kubeadmcontrolplane"
		arg3 := "-A"
		argKubeConfig := "--kubeconfig"
		cmd := exec.Command(prg, arg1, arg2, arg3, argKubeConfig, kubeConfig)
		// Get the result from kubectl and send to Infra Controller
		stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			// log.Fatal(err)
			return
		}

		var getClusterResult []Message
		trimmedString := strings.TrimSpace(string(stdout))
		listTrimmedString := strings.Split(trimmedString, "\n")
		if len(listTrimmedString) < 2 {
			w.Write([]byte(string(stdout)))
		}
		for i, str := range listTrimmedString {
			if i != 0 {
				splitStr := strings.Fields(str)
				msg := Message{splitStr[0], splitStr[1], splitStr[2], splitStr[3]}
				// msgMarshaled, _ := json.Marshal(msg)

				getClusterResult = append(getClusterResult, msg)
			}
		}
		jsongetClusterResult, errorConvertJson := json.Marshal(getClusterResult)
		if errorConvertJson != nil {
			fmt.Println("error:", errorConvertJson)
		}

		w.Write([]byte(string(jsongetClusterResult)))
	})
	r.Get("/getmachines", func(w http.ResponseWriter, r *http.Request) {

		prg := "./kubectl"
		arg1 := "get"
		arg2 := "machines"
		arg3 := "-A"
		argKubeConfig := "--kubeconfig"
		cmd := exec.Command(prg, arg1, arg2, arg3, argKubeConfig, kubeConfig)

		stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			// log.Fatal(err)
			return
		}

		var getMachinesResult []Machine
		trimmedString := strings.TrimSpace(string(stdout))
		listTrimmedString := strings.Split(trimmedString, "\n")

		if len(listTrimmedString) < 2 {
			w.Write([]byte(string(stdout)))
		}
		for i, str := range listTrimmedString {
			if i != 0 {
				splitStr := strings.Fields(str)
				var machineItem Machine
				if len(str) > 6 {
					machineItem = Machine{splitStr[0], splitStr[1], splitStr[2], splitStr[3], splitStr[4], splitStr[5], splitStr[6], splitStr[7]}
				} else {
					machineItem.Namespace = splitStr[0]
					machineItem.Name = splitStr[1]
					machineItem.Cluster = splitStr[2]
					if len(splitStr) > 3 {
						if splitStr[3] == "Failed" || splitStr[3] == "Deleting" || splitStr[3] == "Pending" {
							machineItem.Phase = splitStr[3]
							machineItem.NodeName = "Null"
							machineItem.ProviderID = "Null"
							machineItem.Age = "Null"
							machineItem.Version = "Null"
						} else if len(splitStr) > 4 {
							machineItem.Phase = splitStr[4]
							machineItem.NodeName = "Null"
							machineItem.ProviderID = splitStr[3]
							machineItem.Age = "Null"
							machineItem.Version = "Null"
						}
						// machineItem.NodeName
					}

				}

				// msgMarshaled, _ := json.Marshal(msg)

				getMachinesResult = append(getMachinesResult, machineItem)
			}
		}
		jsonGetMachinesResult, errorConvertJson := json.Marshal(getMachinesResult)
		if errorConvertJson != nil {
			fmt.Println("error:", errorConvertJson)
		}

		w.Write([]byte(string(jsonGetMachinesResult)))
	})
	r.Get("/getkubeconfig", func(w http.ResponseWriter, r *http.Request) {
		var clusterName string
		clusterName = r.Header.Get("clustername")
		if len(clusterName) < 1 {
			fmt.Println("Missing clustername field in request")
		}
		prg := "./clusterctl"
		arg1 := "get"
		arg2 := "kubeconfig"
		// argKubeConfig := "--kubeconfig " + kubeConfig
		cmd := exec.Command(prg, arg1, arg2, clusterName)
		// Get the result from kubectl and send to Infra Controller
		stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			// log.Fatal(err)
			return
		}
		var kubeConfigRaw = KubeConfigMessage{Name: clusterName, KubeConfig: string(stdout)}
		jsongetClusterResult, errorConvertJson := json.Marshal(kubeConfigRaw)
		if errorConvertJson != nil {
			fmt.Println("error when convert JSON", jsongetClusterResult, errorConvertJson)

		}

		w.Write([]byte(string(stdout)))
	})
	r.Get("/testCreateNewCluster", func(w http.ResponseWriter, r *http.Request) {
		runK8sJobs(yamlFileTemplate.JobsTemplate)
		w.Write([]byte(string("Creating k8s Job: ")))
	})
	// Create New cluster in OPENSTACK through call clusterASPI
	r.Post("/createNewCluster", func(w http.ResponseWriter, r *http.Request) {

		// defer r.Body.Close()
		fmt.Println("Received create new Cluster Request")
		httpPostBody, err := ioutil.ReadAll(r.Body) //<--- here!

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(httpPostBody))
		var clusterConfig ClusterRecord
		err = json.Unmarshal(httpPostBody, &clusterConfig)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println((clusterConfig))
		fmt.Println("Before Applying cluster YAML FIle")
		clusterYamlFile, ok := generateClusterYamlFile(clusterConfig)
		fmt.Println("Print a part of yaml file")
		if ok {
			prg := "./kubectl"

			argKubeConfig := "--kubeconfig"
			argCreate := "create"
			namespaceKeyword := "ns"
			// Create namespace and don't care about the result
			fmt.Println("Creating namespace if namespace doesnt exist.....")
			cmdtemp := exec.Command(prg, argCreate, namespaceKeyword, namespaceClusterAPI, argKubeConfig, kubeConfig)
			stdoutTemp1, errTemp := cmdtemp.Output()
			if errTemp != nil {
				fmt.Println("Namespace already Exist")
				fmt.Println(errTemp.Error())
				fmt.Println(string(stdoutTemp1))
				// log.Fatal(err)
			}

			fmt.Println("Created namespace: ", namespaceClusterAPI, "\nOutput command: ", string(stdoutTemp1))
			//------------------------------------------------
			arg1 := "apply"
			arg2 := "-f"
			fmt.Println("Applying cluster template file: ", clusterYamlFile)
			cmd := exec.Command(prg, arg1, arg2, clusterYamlFile, argKubeConfig, kubeConfig)
			// Get the result from kubectl and send to Infra Controller
			fmt.Println("Print command: ", cmd.Path, cmd.Args, cmd.Env)
			stdout1, err := cmd.Output()

			if err != nil {
				fmt.Println("Error applying cluster occurred")
				// Print Error and details of error happend
				fmt.Println(fmt.Sprint(err) + ": " + string(stdout1))
				// log.Fatal(err)
			}

			fmt.Println("Output kubectl apply -f ", string(stdout1))
			listYamlFileClusterAPI = append(listYamlFileClusterAPI, clusterYamlFile)
			// Run k8s Job to waiting for cluster provisioning status, get kubeconfig and register cluster to EMCO
			runK8sJobs(yamlFileTemplate.JobsTemplate)
		}

		w.Write([]byte(string("Creating cluster: ") + clusterConfig.Name))
	})
	// CURL POST Request Example
	// curl -X POST --data-binary @./test.sh http://127.0.0.1:3333/runBashScript
	r.Post("/runBashScript", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Received runBashScript Request")
		httpPostBody, err := ioutil.ReadAll(r.Body) //<--- here!
		fmt.Println(string(httpPostBody))

		// Save content to file
		filePath := saveContentToBashFile(httpPostBody, "bash.sh")
		if filePath == "error" {
			return
		}
		fmt.Println("Print bash file path: ", filePath)
		cmd := exec.Command("/bin/sh", filePath)
		// Run the bash file
		fmt.Println("Print command: ", cmd.Path, cmd.Args, cmd.Env)
		stdout1, err := cmd.Output()

		// prg := "echo " + httpPostBody
		// arg := " | kubectl apply -f -"
		// cmd := exec.Command(prg, arg)
		// stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(stdout1))
			return
		}
		w.Write([]byte(string(stdout1)))
	})
	// Run an K8s Jobs Endpoint
	r.Post("/runk8sjobs", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Received runk8sjobs Request")
		httpPostBody, err := ioutil.ReadAll(r.Body) //<--- here!
		fmt.Println(string(httpPostBody))
		// Replace ENV Var in file
		stringHttpPostBody := string(httpPostBody)
		// Generate job name. env values
		jobName := "job-" + RandomString(6)
		clusterName := jobName                  // "placeholder-cluster-name"
		clusterNamespace := namespaceClusterAPI //"placeholder-cluster-namespace"
		// objectName := "object-name"
		// statusObject := "status-object"

		stringHttpPostBody = strings.Replace(stringHttpPostBody, "placeholder-name", jobName, 1)
		stringHttpPostBody = strings.Replace(stringHttpPostBody, "placeholder-cluster-name", clusterName, 1)
		stringHttpPostBody = strings.Replace(stringHttpPostBody, "placeholder-cluster-namespace", clusterNamespace, 1)
		//
		// Save content to file
		filePath := saveContentToYamlFile(stringHttpPostBody, "jobs.yaml")
		if filePath == "error" {
			return
		}
		fmt.Println("Print bash file path: ", filePath)
		cmd := exec.Command("/bin/sh", filePath)
		// Run the bash file
		fmt.Println("Print command: ", cmd.Path, cmd.Args, cmd.Env)
		stdout1, err := cmd.Output()

		// prg := "echo " + httpPostBody
		// arg := " | kubectl apply -f -"
		// cmd := exec.Command(prg, arg)
		// stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(stdout1))
			return
		}
		w.Write([]byte(string(stdout1)))
	})
	fmt.Println("Start Server at port", serverPort)
	http.ListenAndServe(serverPort, r)
}

// ==============================FUNCTIONS============================
func getEnv(key string, defaultValue string) string {
	fmt.Println("Get Env KUBECONFIG", os.Getenv("KUBECONFIG"))
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func createTempFolder(nameCluster string) string {
	dname, err := os.MkdirTemp("", nameCluster)
	if err != nil {
		panic(err)
	}
	return dname
}
func generateClusterYamlFile(record ClusterRecord) (string, bool) {
	fmt.Println("Generate Cluster Yaml File", record.Name)
	arg := "./clusterctl"
	arg1 := "generate"
	arg2 := "cluster"
	// clusterctl generate cluster capi-quickstart   --kubernetes-version v1.23.5   --control-plane-machine-count=3   --worker-machine-count=3   > capi-quickstart.yaml
	argK8sVersion := "--kubernetes-version"
	argControlPlaneMachineCount := "--control-plane-machine-count"
	argWorkerMachineCount := "--worker-machine-count"
	argTargetNamespace := "--target-namespace"
	// arg3 := "> /tmp/a.yaml"
	cmd := exec.Command(arg, arg1, arg2, record.Name, argK8sVersion, record.KubernetesVersion, argControlPlaneMachineCount, record.ControlPlaneMachineCount, argWorkerMachineCount, record.KubernetesMachineCount, argTargetNamespace, namespaceClusterAPI)
	fmt.Println("Print command: ", cmd.Path, cmd.Args, cmd.Env)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println("Error occurred")
		fmt.Println(err.Error())
		return string(stdout), false
	}
	fmt.Println("stdout", string(stdout))
	// Create folder
	// And write to yaml file
	tempFolder := createTempFolder(record.Name)
	fmt.Println("Create  temp folder", tempFolder)
	templateClusterFile := filepath.Join(tempFolder, record.Name)
	fmt.Println("Create  temp file", templateClusterFile)
	err = os.WriteFile(templateClusterFile, stdout, 0777)
	fmt.Println("Write  temp file", templateClusterFile)
	if err != nil {
		fmt.Println(err.Error())
		// log.Fatal(err)
		return "error", false
	}
	// Replace CDIR block
	// sed 's/192.168.0.0/10.244.0.0/' templateClusterFile
	cmd1 := exec.Command("sed", "-i", "s/192.168.0.0/10.244.0.0/", templateClusterFile)
	stdout2, err2 := cmd1.Output()
	if err2 != nil {
		fmt.Println("Error occurred when replace CIDR block")
		fmt.Println("stdout2", string(stdout2))
		fmt.Println(err.Error())
		return string(stdout2), false
	}
	// Add CNI meta data
	fmt.Println("Replace CNI label...")
	cmd2 := exec.Command("sed", "-i", "0,/spec:/s//  labels:\n    cni: flannel\nspec:/", templateClusterFile)
	fmt.Println("Print command: ", cmd2.Path, cmd2.Args, cmd2.Env)
	stdout3, err3 := cmd2.Output()
	if err3 != nil {
		fmt.Println("Error occurred when add CNI label")
		fmt.Println("stdout3", string(stdout3))
		return string(stdout3), false
	}
	fmt.Println("Print file after replace CNI.. ")
	content, err := ioutil.ReadFile(templateClusterFile)

	if err != nil {
		fmt.Println("Cant open file")
	} else {
		fmt.Println(string(content))
	}

	//
	return templateClusterFile, true
}
func Command(name string, arg ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err != nil {
			// cmd.lookPathErr  = err
			fmt.Println("Error lookpath")
		} else {
			cmd.Path = lp
		}
	}
	return cmd
}

// func generateMachineControlPlaneHealthCheck(clusterName string) string {
// 	return fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1beta1
// 	kind: MachineHealthCheck
// 	metadata:
// 	  name: %s-unhealthy-controlplane
// 	spec:
// 	  clusterName: %s
// 	  maxUnhealthy: 100%
// 	  selector:
// 		matchLabels:
// 		  cluster.x-k8s.io/control-plane: ""
// 	  unhealthyConditions:
// 		- type: Ready
// 		  status: Unknown
// 		  timeout: 1s
// 	`, clusterName, clusterName)
// }

// func generateMachineWorkerHealthCheck(clusterName string) string {
// 	return fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1beta1
// 	kind: MachineHealthCheck
// 	metadata:
// 	  name: %s-unhealthy
// 	spec:
// 	  clusterName: %s
// 	  maxUnhealthy: 100%
// 	  nodeStartupTimeout: 10m
// 	  selector:
// 		matchLabels:
// 		  cluster.x-k8s.io/deployment-name: %s-md-0
// 	  unhealthyConditions:
// 		- type: Ready
// 		  status: Unknown
// 		  timeout: 1s
// 	`, clusterName, clusterName, clusterName)
// }
// func createCNIFlannelPlugin() string {

// 	return string(`apiVersion: addons.cluster.x-k8s.io/v1alpha3
// 	kind: ClusterResourceSet
// 	metadata:
// 	  name: cni-flannel
// 	spec:
// 	  clusterSelector:
// 		matchLabels:
// 		  cni: flannel
// 	  resources:
// 	  - kind: ConfigMap
// 		name: flannel-configmap`)
// }

//	func addCNILabelToYamlFile(yamlFile string) string {
//		labelCNI := "\n  labels:\n    cni: flannel\n"
//		strings.Index
//		return finalYamlFile
//	}
// func addCNILabelToYamlFile(yamlFile string) string {
// 	// labelCNI := "\n  labels:\n    cni: flannel\n"
// 	// strings.Index()

//		return yamlFile
//	}
func saveContentToBashFile(content []byte, fileName string) string {
	// var fileName string
	tempFolder := createTempFolder(fileName)

	bashFilePath := filepath.Join(tempFolder, fileName)

	fmt.Println("Write  bash file", bashFilePath)
	// Check is sh file include #!/bin/sh part
	contentStr := string(content)
	var err error
	if strings.Contains(contentStr, `#!/bin/sh`) {
		err = os.WriteFile(bashFilePath, content, 0777)
	} else {
		contentStr = `#!/bin/sh` + "\n" + contentStr
		err = os.WriteFile(bashFilePath, []byte(contentStr), 0777)
	}

	if err != nil {
		fmt.Println(err.Error())
		return "error"
	}

	return bashFilePath
}

// Save yaml file
func saveContentToYamlFile(content string, fileName string) string {
	// var fileName string
	tempFolder := createTempFolder(fileName)

	bashFilePath := filepath.Join(tempFolder, fileName)

	fmt.Println("Write  yaml file", bashFilePath)
	// Check is sh file include #!/bin/sh part
	// contentStr := string(content)
	var err error

	err = os.WriteFile(bashFilePath, []byte(content), 0777)

	if err != nil {
		fmt.Println(err.Error())
		return "error"
	}

	return bashFilePath
}

func getAndParseNamespaceForCLusterApi() string {
	var namespaceClusterApi string
	cloudYamlB64 := getEnv("OPENSTACK_CLOUD_YAML_B64", "default")
	data, err := base64.StdEncoding.DecodeString(cloudYamlB64)
	if err != nil {
		fmt.Println("error decode 64:", err)
		return "default"
	}
	cloudYaml := CloudYaml{}
	err = yaml.Unmarshal([]byte(data), &cloudYaml)
	if err != nil {
		fmt.Println("error read yaml file:", err)
		namespaceClusterApi = "default"
		fmt.Println("Name space for cluster API is assign to default value: ", namespaceClusterApi)
	} else {
		cloudProviderName := "openstack"
		cloudName := maps.Keys(cloudYaml.Clouds)[0]
		namespaceClusterApi = cloudProviderName + "-" + cloudName + "-" + cloudYaml.Clouds[cloudName].AuthInform.ProjectName
	}

	return namespaceClusterApi
}

func runK8sJobs(yamlTemplate string) {
	fmt.Println("Begin runk8sjobs Request")
	// httpPostBody, err := ioutil.ReadAll(r.Body) //<--- here!
	// fmt.Println(string(httpPostBody))
	// Replace ENV Var in file
	stringHttpPostBody := yamlFileTemplate.JobsTemplate //string(httpPostBody)
	// Generate job name. env values
	jobName := "job-" + RandomString(6)
	clusterName := jobName                  // "placeholder-cluster-name"
	clusterNamespace := namespaceClusterAPI //"placeholder-cluster-namespace"
	// objectName := "object-name"
	// statusObject := "status-object"

	stringHttpPostBody = strings.Replace(stringHttpPostBody, "placeholder-name", jobName, 1)
	stringHttpPostBody = strings.Replace(stringHttpPostBody, "placeholder-cluster-name", clusterName, 1)
	stringHttpPostBody = strings.Replace(stringHttpPostBody, "placeholder-cluster-namespace", clusterNamespace, 1)
	//
	// Save content to file
	filePath := saveContentToYamlFile(stringHttpPostBody, "jobs")
	if filePath == "error" {
		fmt.Println("Error when save content to yaml file")
		return
	}
	fmt.Println("Print k8s Jobs path: ", filePath)
	prg := "./kubectl"

	argKubeConfig := "--kubeconfig"
	arg1 := "apply"
	arg2 := "-f"
	fmt.Println("Applying k8s Job  file: ", stringHttpPostBody, "\n------------------------------\n")
	cmd := exec.Command(prg, arg1, arg2, filePath, argKubeConfig, kubeConfig)
	// Get the result from kubectl and send to Infra Controller
	fmt.Println("Print command: ", cmd.Path, cmd.Args, cmd.Env)
	stdout1, err := cmd.Output()

	if err != nil {
		fmt.Println("Error applying K8s Jobs occurred")
		// Print Error and details of error happend
		fmt.Println(fmt.Sprint(err) + ": " + string(stdout1))
		// log.Fatal(err)
	}

	fmt.Println("Output kubectl apply -f ", string(stdout1))
	// stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(stdout1))
		return
	}
}
