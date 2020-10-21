package main

import (
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
	restclient "k8s.io/client-go/rest"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/api/errors"
	clusterquotav1client "github.com/openshift/client-go/quota/clientset/versioned/typed/quota/v1"
	clusterquotav1 "github.com/openshift/api/quota/v1"
	apiv1 "k8s.io/api/core/v1"
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

//TODO

func checkIfProjectQuotaExists(projectName string , quotaName string, clientset *kubernetes.Clientset) bool {
	_, err := clientset.CoreV1().ResourceQuotas(projectName).Get(quotaName, v1.GetOptions{})
	if err != nil {
		log.Debug("Project quota for " + projectName + " does no exist")
		return false
	} else {
		return true
	}
}

func checkIfClusterQuotaExists(clusterQuotaName string, config *restclient.Config ) bool {
	log.Debug("Checking if cluster quota " + clusterQuotaName + " exists")
	quotaSet,_ := clusterquotav1client.NewForConfig(config)
	clusterQuotaInfo,_ := quotaSet.ClusterResourceQuotas().Get(clusterQuotaName, v1.GetOptions{})
	if (clusterQuotaInfo.Spec.Quota.Size() == 0) {
		log.Debug("Ni ma")
		return false
	} else {
		log.Debug("Quota " + clusterQuotaName + " exists")
		return true
	}
}