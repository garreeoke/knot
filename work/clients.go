package work

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"

)

func (k *Knot) GetK8Client() error {

	var config *rest.Config
	var err error
	if k.Auth == OnCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	} else if k.Auth == Local {
		var kubeconfig *string
		if home := homeDir(); home != "" {
			filepath.Join()
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("",*kubeconfig)
		if err != nil {
			return err
		}
	}
	// Create client set
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	k.Client = clientset

	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		log.Println("HOME: ", h)
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
