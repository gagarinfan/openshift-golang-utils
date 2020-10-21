package main

import (
	//"fmt"
	"net"
	"io/ioutil"
	"strings"
	"gopkg.in/yaml.v2"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	//openshift "github.com/openshift/api/network/v1"
	//"github.com/openshift/client-go/network/clientset/versioned/typed/network/v1"
	//projectv1client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	networkv1client "github.com/openshift/client-go/network/clientset/versioned/typed/network/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//clusterquotav1client "github.com/openshift/client-go/quota/clientset/versioned/typed/quota/v1"
)
type configYaml struct {
	Wit  string
	Bsi  string
	AdminGroup string `yaml:"adminGroup"`
	Clusterquota string `yaml:"clusterquota"`
  Envs []Env
}

type Env struct {
	Name   string
	CPU    string
	Memory string
	Egress []string
	Clusterquota string `yaml:"clusterquota"`
	Ingress_allowed []struct {
		Name   string
	}
	Custom_objects []struct{
		Name   string
	}
	Labels []struct {
		Name string
		Value string
	}
}

func IsIpv4Net(host string) bool {
	return net.ParseIP(host) != nil
}

func validateAddresses(ips []string, envName string) int{
	var wrongAddresses []string
	for _, ip := range ips {
		if !IsIpv4Net(ip) { wrongAddresses = append(wrongAddresses, ip) }
	}
	if len(wrongAddresses) > 0 {
		log.Warn("Provided IPs for " + envName + " are not valid IPv4 format: " + strings.Join(wrongAddresses,", "))
		return 1
	} else {
		return 0
	}
}

func loginToKubernetes()(*networkv1client.NetworkV1Client){

	config, err := clientcmd.BuildConfigFromFlags("", "kubeconfig")

	if err != nil {
		panic(err.Error())
	}

	networkset, err := networkv1client.NewForConfig(config)

	_, error := networkset.NetNamespaces().List(v1.ListOptions{})
	
	if errors.IsUnauthorized(error) {
		log.Error("Problem with authorization")
	}

	return networkset
}

func createEgressIP(addresses []string, namespace string, networkset *networkv1client.NetworkV1Client) {
	netnamespace, err := networkset.NetNamespaces().Get(namespace, v1.GetOptions{})
	if err != nil {
		log.Fatal("Could not find netnamespace for " + namespace + ". Caused by: " + err.Error())
	}
	netnamespace.EgressIPs = addresses
	_, errUpdate := networkset.NetNamespaces().Update(netnamespace)
	if errUpdate != nil {
		log.Fatal("Could not assign egress IP addresses: " + strings.Join(addresses,", ") + " to netnamsepace " + namespace + ". Caused by: " + errUpdate.Error())
	}
	log.Info("Assigned egress IP addresses " + strings.Join(addresses,", ") + " to netnamespace " + namespace)
}

func main() {
	log.SetLevel(log.DebugLevel)
	files, _ := ioutil.ReadDir("./data/")
	var errValue int

	for _, f := range files {
		dat, _ := ioutil.ReadFile("data/" + f.Name() + "/main.yaml.kv")
		var projectCfg configYaml
		yaml.Unmarshal(dat, &projectCfg)
		log.Debug(projectCfg)
		for _, envName := range projectCfg.Envs {
			if (len(envName.Egress) != 0) {
				errValue += validateAddresses(envName.Egress, envName.Name)
			}
		}
		if errValue > 0 {log.Fatal("EgressIP addresses validation has failed!")}
		for _, envName := range projectCfg.Envs {
			if envName.Name == "prod_dmz" {
				//do wywalenia
				networkset := loginToKubernetes()
				//end
				createEgressIP(envName.Egress, f.Name(), networkset)
			}
		}
	}
}