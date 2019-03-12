package common

import (

	"encoding/json"
	"encoding/base64"
	"fmt"

	"applariat.io/cluster-manager/types"
	"applariat.io/propeller/kube"
	propeller "applariat.io/propeller/types"

	"os"
	"time"
)

// addOns will add/delete the appLariat addons to the cluster
func addOns(action string, cd *types.ClusterData) error {

	var msg string
	k8, err := createK8(cd)
	if err != nil {
		msg = fmt.Sprintf("Add-ons %v failed: %v", action, err)
		cd.Job.Log.Println(msg)
		return err
	}

	if k8.Location.Status.State != propeller.AplLocDeployFailed {
		clusterSvcStatus := propeller.ClusterSvcStatus{
			ConfigMap: true,
			Events: true,
			Action: action,
		}
		toInstall := []string{"registry", "propeller", "policy", "monitoring", "autoscaler", "storageclasses"}
		if k8.Location.Type == types.ProviderVSphere {
			toInstall = []string{"propeller", "policy"}
			if v,ok := cd.Location.Annotations["vke"]; ok {
				if v == "true" {
					// Install cluster bound registry
					//toInstall = append(toInstall, "registry")
				}
			} else {
				toInstall = append(toInstall, "storageclasses")
			}
		}
		if len(cd.Location.Services) > 0 {
			toInstall = append(toInstall, cd.Location.Services...)
		}
		clusterSvcStatus.AddOns = toInstall
		clusterSvcStatus.Status = make(chan propeller.DeploymentStatus, len(clusterSvcStatus.AddOns))
		clusterSvcStatus.Failures = make(chan int, len(clusterSvcStatus.AddOns))
		cd.ClusterSvcFailures = clusterSvcStatus.Failures

		if k8.Location.DNS.Enabled {
			clusterSvcStatus.AddOns = append(clusterSvcStatus.AddOns, "ext-dns")
		}

		go k8.AddOns(&clusterSvcStatus)

		// Loop on the channel here and process each service ... sending the appropriate event
		for depStatus := range clusterSvcStatus.Status  {
				for _, componentStatus := range depStatus.Components {
					if componentStatus.State != propeller.AplComponentRunning {
						clusterSvcStatus.Failures <- 1
					}
					// Send event for cluster service
					err = cd.UpdateClusterSvcStatus(&depStatus)
					if err != nil {
						cd.Job.Log.Println("Unable to send event to apl: ", err)
					}
				}
		}
		close(clusterSvcStatus.Failures)

		// Update the location
		if action == propeller.AddOnCreate {
			// Update location object on cd
			err = cd.UpdateLocationFromMgr(k8.Location)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func locSecret(action string, cd *types.ClusterData) error {

	cd.Job.Log.Println("Creating new locSecret")
	k8, err := createK8(cd)
	if err != nil {
		return err
	}

	loc, err := json.Marshal(cd.Location)
	if err != nil {
		return err
	}

	aplLocSecret := propeller.AplLocSecret + "-" + cd.Location.ID
	secretMap := map[string][]byte{
		aplLocSecret: loc,
	}
	switch action {
	case "create":
		success := false
		for tries := 1; !success && tries <= 5; tries++ {
			err = k8.NewSecret(aplLocSecret, "default", secretMap)
			if err == nil {
				success = true
			} else if err != nil && err.Error() == "net/http: TLS handshake timeout" {
				cd.Job.Log.Println("Retrying due to TLS timeout")
				time.Sleep(2 * time.Second)
			}
		}
	case "delete":
		err = k8.DeleteSecret([]string{aplLocSecret})
	}

	if err != nil {
		return err
	}

	return nil
}

func addConfig(cd *types.ClusterData) error {

	cd.Job.Log.Println("Building config for cluster")
	k8, err := createK8(cd)
	if err != nil {
		return err
	}
	k8.Name = cd.Location.Name

	config, err := k8.BuildKubeCfg()
	if err != nil {
		return err
	}

	cfg, err := json.Marshal(config)
	if err != nil {
		return err
	}

	x := []byte(string(cfg))
	if os.Getenv("APL_STANDALONE") == "true" {
		cd.Job.Log.Println("Kube_Config: ", string(x))
	}

	k8.Location.Cluster.Config = base64.RawStdEncoding.EncodeToString(x)
	err = cd.UpdateLocationFromMgr(k8.Location)
	if err != nil {
		return err
	}

	return nil
}

func createK8(cd *types.ClusterData) (kube.K8, error) {

	var k8 kube.K8
	var err error

	namespace := "kube-system"
	k8.DeployID = namespace
	k8.Name = namespace
	k8.Location = *cd.Location

	if k8.Location.Type == propeller.ProviderVsphere || k8.Location.Type == propeller.ProviderVKE {
		k8.Name = "default"
	}

	k8.Log = cd.Job.Log
	k8.OnCluster = cd.OnCluster

	if k8.OnCluster {
		err = k8.Auth(false)
	} else {
		err = k8.Auth(true)
	}

	if err != nil {
		k8.Log.Println("AUTH_ERR: ", err)
		return k8, err
	}

	return k8, nil
}


