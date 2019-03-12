package common

import (
	"applariat.io/cluster-manager/types"
	propeller "applariat.io/propeller/types"
	"errors"
	"fmt"
)

func Action(cd *types.ClusterData, mgr types.ClusterManager) error {

	var err error
	switch cd.Job.Action {
	case "create":
		cd.ClusterExists = false
		msg := "Provisioning"
		if !cd.Location.AplManaged {
			msg = "Importing"
		}
		cd.Status <- propeller.ClusterStatus{
			Description: msg + " cluster",
			Percent:     1,
			State:       "creating",
		}

		err = mgr.Create(cd)
		if err != nil {
			cd.Job.Log.Println("Action error: ", err)
			return err
		}

		cd.Status <- propeller.ClusterStatus{
			Description: "Configuring and verifying cluster",
			Percent:     1,
			State:       propeller.AplLocDeployLocked,
		}

		mgr.Result(cd)
		err = cd.UpdateLocationFromMgr(mgr)
		if err != nil {
			cd.Job.Log.Println("Error updating loc_deploy in cd object: ", err)
			return err
		}

		// If we get here, update the running nodes and failures should not return error
		cd.Status <- propeller.ClusterStatus{
			Description:  fmt.Sprintf("Nodes running: %v", cd.ExpectedNodes),
			RunningNodes: cd.ExpectedNodes,
		}

		if !cd.Location.AplOnApl {
			err = addOns(propeller.AddOnCreate, cd)
			descr := "waiting"
			if err != nil {
				descr = "Cluster services did not install: " + err.Error()
			} else if len(cd.ClusterSvcFailures) > 0 {
				descr = "One or more cluster services did not install"
			}
			if descr != "waiting" {
				cd.Status <- propeller.ClusterStatus{
					Description: descr,
					Percent:     1,
					State:       propeller.AplLocDeployUnavailable,
				}
			}
		}

		// Adding secret
		err = locSecret("create", cd)
		if err != nil {
			cd.Status <- propeller.ClusterStatus{
				State:       propeller.AplLocDeployUnavailable,
				Description: "Secret Configuration failed: " + err.Error(),
			}
		}

		// Add kubeconfig to cluster record
		err = addConfig(cd)
		if err != nil {
			cd.Status <- propeller.ClusterStatus{
				Description: "KubeConfig creation failed failed: " + err.Error(),
			}
		}

		if cd.Location.Status.State != propeller.AplLocDeployFailed && cd.Location.Status.State != propeller.AplLocDeployUnavailable {
			cd.Status <- propeller.ClusterStatus{
				State:       propeller.AplLocDeployAvailable,
				Description: "Cluster is available",
				Percent:     100,
			}
		} else {
			cd.Status <- propeller.ClusterStatus {
				Percent:     100,
			}
		}

	case "delete":
		cd.ClusterExists = true
		cd.Status <- propeller.ClusterStatus{
			Description: "Deleting existing deployments and services",
			Percent:     1,
			State:       "deleting",
		}
		k8, err := createK8(cd)
		if !cd.Location.AplOnApl {
			err = k8.PreClusterDelete()
			cd.Status <- propeller.ClusterStatus{
				Description: "Deleting cluster",
				Percent:     1,
			}
		}
		// Remove add-ons if not apl-managed
		if !cd.Location.AplManaged {
			if !cd.Location.AplOnApl {
				err = addOns(propeller.AddOnDelete, cd)
				if err != nil {
					cd.Status <- propeller.ClusterStatus{
						Description: err.Error(),
						Percent:     1,
						State:       propeller.AplLocDeployUnavailable,
					}
					return err
				}
			}
		} else {
			if !cd.Location.AplOnApl {
				err = mgr.Destroy(cd)
				if err != nil {
					cd.Status <- propeller.ClusterStatus{
						Description: err.Error(),
						Percent:     100,
						State:       propeller.AplLocDeployUnavailable,
					}
					return err
				}
			} else if cd.Location.AplOnApl{
				// Delete the secret
				err = locSecret("delete", cd)
				if err != nil {
					cd.Job.Log.Println("Unable to delete secret:", err)
				}
			}
		}

		mgr.Result(cd)
		if cd.Location.Status.State != propeller.AplLocDeployUnavailable {
			cd.ClusterExists = false
			cd.Status <- propeller.ClusterStatus{
				State:   propeller.AplLocDeployDeleted,
				Percent: 100,
			}
		}
		err = cd.UpdateLocationFromMgr(mgr)
		if err != nil {
			cd.Job.Log.Println("Error updating loc_deploy in cd object: ", err)
		}
	case "upgrade":
		err = mgr.Upgrade(cd)
	default:
		err = errors.New("Invalid CLOUD_ACTION")
		return err
	}

	return nil
}
