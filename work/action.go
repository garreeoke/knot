package work

import (
	"encoding/json"
	"github.com/garreeoke/kates"
	"github.com/ghodss/yaml"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	netv1beta "k8s.io/api/networking/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"log"
	"os"
	"strings"
	"time"
)

const (
	create = "create"
)

func (k *Knot) AddOns() error {

	workDir, err := os.Open(k.WorkDir)
	if err != nil {
		return err
	}
	defer workDir.Close()
	// Get a listing of the directory, then process each one
	workDirList, err := workDir.Readdir(0)
	if err != nil {
		return err
	}

	doneCh := make(chan int, len(workDirList))
	errorCh := make(chan string)
	outputCh := make(chan *kates.Output)

	// Loop on the error channel and just print them for logging
	go func() {
		for msg := range errorCh {
			log.Println(msg)
		}
	}()

	go func() {
		for output := range outputCh {
			k.Output = append(k.Output, output)
		}
	}()

	// Look in work directory for files and directories
	for _, workDirFileHandle := range workDirList {
		log.Println("File name: ", workDirFileHandle.Name() )
		if !workDirFileHandle.IsDir() {
			go func(workDirFile os.FileInfo){
				if len(strings.Split(workDirFile.Name(),".yaml")) == 2 {
					log.Println("Posting: ", workDirFile.Name() )
					err = k.postAddOn(workDirFile, k.WorkDir+"/"+workDirFile.Name(), outputCh)
					if err != nil {
						log.Println(err)
						errorCh <- k.WorkDir + "/" + workDirFile.Name() + " ERROR: " + err.Error()
					}
				}
				doneCh <- 1
			}(workDirFileHandle) // End of go func for files
		} else if workDirFileHandle.IsDir() {
			// Channels here ?
			go func(workDirFile os.FileInfo) {
				// Check if directory
				subDirPath := k.WorkDir + "/" + workDirFile.Name()
				// Open sub directory
				subDir, err := os.Open(subDirPath)
				if err != nil {
					errorCh <- subDirPath + " ERROR: " + err.Error()
				} else {
					subDirList, err := subDir.Readdir(0)
					if err != nil {
						errorCh <- subDirPath + " ERROR: " + err.Error()
					} else {
						for _, subDirFileInfo := range subDirList {
							subDirFilePath := subDirPath + "/" + subDirFileInfo.Name()
							if !subDirFileInfo.IsDir() {
								if len(strings.Split(subDirFileInfo.Name(),".yaml")) == 2 {
									log.Println("Posting: ", subDirFileInfo.Name() )
									err = k.postAddOn(subDirFileInfo, subDirFilePath, outputCh)
									if err != nil {
										errorCh <- subDirPath + "/" + subDirFileInfo.Name() + " ERROR: " + err.Error()
									}
								}
							}
						}
					}
				}
				_ = subDir.Close()
				doneCh <- 1
			}(workDirFileHandle) // End of go func for directories
		}
	}

	// Try for three minutes ... could just do endless loop here?
	for tries := 1; len(doneCh) < len(workDirList); tries++ {
		if tries <= 180 {
			time.Sleep(time.Duration(1000) * time.Millisecond)
		} else {
			break
		}
	}
	close(errorCh)
	close(doneCh)
	close(outputCh)

	return nil
}

// postAddOn ... does work of modifying and posting of add ons
func (k *Knot) postAddOn(file os.FileInfo, fullPath string, output chan *kates.Output) error {

	fileData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// Put contents of file into a map
	var i interface{}
	yToJ, err := yaml.YAMLToJSON(fileData)
	err = json.Unmarshal(yToJ, &i)
	if err != nil {
		return err
	}
	m := i.(map[string]interface{})
	katesInput := kates.Input{
		Client: k.Client,
	}
	var katesOutput *kates.Output

	// Add annotation
	switch m["kind"] {
	case K8KindConfigMap:
		cm := &apiv1.ConfigMap{}
		err := json.Unmarshal(yToJ, &cm)
		if err != nil {
			return err
		}
		katesInput.Data = cm
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateConfigMap(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindStorageClass:
		sc := &storagev1.StorageClass{}
		err = json.Unmarshal(yToJ, &sc)
		if err != nil {
			return err
		}
		katesInput.Data = sc
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateStorageClass(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindDeployment:
		dep := &appsv1.Deployment{}
		err = json.Unmarshal(yToJ, &dep)
		if err != nil {
			return err
		}
		katesInput.Data = dep
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateDeployment(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindStatefulSet:
		ss := &appsv1.StatefulSet{}
		err = json.Unmarshal(yToJ, &ss)
		if err != nil {
			return err
		}
		katesInput.Data = ss
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateDeployment(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindIngress:
		ing := &netv1beta.Ingress{}
		err = json.Unmarshal(yToJ, &ing)
		if err != nil {
			return err
		}
		katesInput.Data = ing
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateIngress(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindService:
		svc := &apiv1.Service{}
		err = json.Unmarshal(yToJ, &svc)
		if err != nil {
			return err
		}
		katesInput.Data = svc
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateService(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindReplicationController:
		rc := &apiv1.ReplicationController{}
		err = json.Unmarshal(yToJ, &rc)
		if err != nil {
			return err
		}
		katesInput.Data = rc
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateReplicationController(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindDaemonSet:
		ds := &appsv1.DaemonSet{}
		err = json.Unmarshal(yToJ, &ds)
		if err != nil {
			return err
		}
		katesInput.Data = ds
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateDaemonSet(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindServiceAccount:
		sc := &apiv1.ServiceAccount{}
		err = json.Unmarshal(yToJ, &sc)
		if err != nil {
			return err
		}
		katesInput.Data = sc
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateServiceAccount(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindClusterRole:
		cr := &rbacv1.ClusterRole{}
		err = json.Unmarshal(yToJ, &cr)
		if err != nil {
			return err
		}
		katesInput.Data = cr
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateClusterRoles(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindClusterRoleBinding:
		crb := &rbacv1.ClusterRoleBinding{}
		err = json.Unmarshal(yToJ, &crb)
		if err != nil {
			return err
		}
		katesInput.Data = crb
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateClusterRoleBindings(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindRole:
		r := &rbacv1.Role{}
		err = json.Unmarshal(yToJ, &r)
		if err != nil {
			return err
		}
		katesInput.Data = r
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateRole(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindRoldBinding:
		rb := &rbacv1.ClusterRoleBinding{}
		err = json.Unmarshal(yToJ, &rb)
		if err != nil {
			return err
		}
		katesInput.Data = rb
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateRoleBindings(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindJob:
		j := &batchv1.Job{}
		err = json.Unmarshal(yToJ, &j)
		if err != nil {
			return err
		}
		katesInput.Data = j
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateJob(&katesInput)
			if err != nil {
				return err
			}
		}
	case K8KindCronJob:
		cj := &batchv1beta1.CronJob{}
		err = json.Unmarshal(yToJ, &cj)
		if err != nil {
			return err
		}
		katesInput.Data = cj
		switch k.Action {
		case create:
			katesOutput, err = kates.CreateCronJob(&katesInput)
			if err != nil {
				return err
			}
		}
	default:
		log.Println("No matching kind: ", m["kind"])
	}

	if katesOutput != nil {
		output <- katesOutput
	}
	return nil
}

/*else if subDirFileInfo.IsDir() {
								// Open the directory
								subSubDir, err := os.Open(subDirFilePath)
								if err != nil {
									workDirErrCh <- 1
								} else {
									subSubDirList, err := subSubDir.Readdir(0)
									if err != nil {
										workDirErrCh <- 1
									} else {
										for _, subSubDirFileInfo := range subSubDirList {
											subSubFilePath := subDirFilePath + "/" + subSubDirFileInfo.Name()
											if !subSubDirFileInfo.IsDir() {
												err = k.postAddOn(subSubDirFileInfo, subSubFilePath)
												if err != nil {
													workDirErrCh <- 1
												}
											}
										}
									}
									_ = subSubDir.Close()
								}
							}*/
