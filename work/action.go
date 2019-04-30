package work

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garreeoke/kates"
	"github.com/ghodss/yaml"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	netv1beta "k8s.io/api/networking/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"log"
	"os"
	"strings"
	"time"
)

const (
	create         = "create"
	createOrModify = "createOrModify"
	modify         = "modify"
)

func (k *Knot) Tie() error {

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
		go func(workDirFile os.FileInfo) {
			log.Println("File name: ", workDirFile.Name())
			if !workDirFile.IsDir() {
				if len(k.WhiteList) == 0 {
					if len(strings.Split(workDirFile.Name(), ".yaml")) == 2 {
						log.Println("Posting: ", workDirFile.Name())
						err = k.postTie(workDirFile, k.WorkDir+"/"+workDirFile.Name(), outputCh)
						if err != nil {
							log.Println(err)
							errorCh <- k.WorkDir + "/" + workDirFile.Name() + " ERROR: " + err.Error()
						}
					}
				} else {
					log.Println("Skipping: ", workDirFile.Name())
				}
			} else if workDirFile.IsDir() {
				process := false
				if len(k.WhiteList) > 0 {
					for _, w := range k.WhiteList {
						if w == workDirFile.Name() {
							process = true
							break
						}
					}
				} else {
					process = true
				}
				if process {
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
							subDirDoneCh := make(chan int, len(subDirList))
							for _, subDirFileHandle := range subDirList {
								go func(subDirFileInfo os.FileInfo) {
									subDirFilePath := subDirPath + "/" + subDirFileInfo.Name()
									if !subDirFileInfo.IsDir() {
										if len(strings.Split(subDirFileInfo.Name(), ".yaml")) == 2 {
											log.Println("Posting: ", subDirFileInfo.Name())
											err = k.postTie(subDirFileInfo, subDirFilePath, outputCh)
											if err != nil {
												errorCh <- subDirPath + "/" + subDirFileInfo.Name() + " ERROR: " + err.Error()
											}
										}
									}
									subDirDoneCh <- 1
								}(subDirFileHandle)
							}
							for tries := 1; len(subDirDoneCh) < len(subDirList); tries++ {
								if tries <= 180 {
									time.Sleep(time.Duration(1000) * time.Millisecond)
								} else {
									break
								}
							}

						}
					}
					_ = subDir.Close()
				} else {
					log.Println("Skipping: ", workDirFile.Name())
				}
			}
			doneCh <- 1
		}(workDirFileHandle)
	}
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

// postTie ...
func (k *Knot) postTie(file os.FileInfo, fullPath string, output chan *kates.Output) error {

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
		Client:    k.Client,
		Operation: k.Operation,
	}
	var katesOutput *kates.Output

	// Add annotation
	for retries := 0; retries <= 30; retries++ {
		retry := true
		switch m["kind"] {
		case K8KindConfigMap:
			cm := &apiv1.ConfigMap{}
			err = json.Unmarshal(yToJ, &cm)
			if err != nil {
				return err
			}
			katesInput.Data = cm
			katesOutput, err = kates.ConfigMap(&katesInput)
		case K8KindCronJob:
			cj := &batchv1beta1.CronJob{}
			err = json.Unmarshal(yToJ, &cj)
			if err != nil {
				return err
			}
			katesInput.Data = cj
			katesOutput, err = kates.CronJob(&katesInput)
		case K8KindDaemonSet:
			ds := &appsv1.DaemonSet{}
			err = json.Unmarshal(yToJ, &ds)
			if err != nil {
				return err
			}
			katesInput.Data = ds
			katesOutput, err = kates.DaemonSet(&katesInput)
		case K8KindDeployment:
			dep := &appsv1.Deployment{}
			err = json.Unmarshal(yToJ, &dep)
			if err != nil {
				return err
			}
			katesInput.Data = dep
			katesOutput, err = kates.Deployment(&katesInput)
		case K8KindIngress:
			ing := &netv1beta.Ingress{}
			err = json.Unmarshal(yToJ, &ing)
			if err != nil {
				return err
			}
			katesInput.Data = ing
			katesOutput, err = kates.Ingress(&katesInput)
		case K8KindJob:
			j := &batchv1.Job{}
			err = json.Unmarshal(yToJ, &j)
			if err != nil {
				return err
			}
			katesInput.Data = j
			katesOutput, err = kates.Job(&katesInput)
		case K8KindNamespace:
			ns := &apiv1.Namespace{}
			err = json.Unmarshal(yToJ, &ns)
			if err != nil {
				return err
			}
			katesInput.Data = ns
			katesOutput, err = kates.Namespace(&katesInput)
		case K8KindNetworkPolicy:
			ns := &netv1.NetworkPolicy{}
			err = json.Unmarshal(yToJ, &ns)
			if err != nil {
				return err
			}
			katesInput.Data = ns
			katesOutput, err = kates.NetworkPolicy(&katesInput)
		case K8KindReplicationController:
			rc := &apiv1.ReplicationController{}
			err = json.Unmarshal(yToJ, &rc)
			if err != nil {
				return err
			}
			katesInput.Data = rc
			katesOutput, err = kates.ReplicationController(&katesInput)
		case K8KindClusterRole:
			cr := &rbacv1.ClusterRole{}
			err = json.Unmarshal(yToJ, &cr)
			if err != nil {
				return err
			}
			katesInput.Data = cr
			retry = false
			katesOutput, err = kates.ClusterRoles(&katesInput)
		case K8KindClusterRoleBinding:
			crb := &rbacv1.ClusterRoleBinding{}
			err = json.Unmarshal(yToJ, &crb)
			if err != nil {
				return err
			}
			katesInput.Data = crb
			retry = false
			katesOutput, err = kates.ClusterRoleBindings(&katesInput)
		case K8KindRole:
			r := &rbacv1.Role{}
			err = json.Unmarshal(yToJ, &r)
			if err != nil {
				return err
			}
			katesInput.Data = r
			katesOutput, err = kates.Role(&katesInput)
		case K8KindRoldBinding:
			rb := &rbacv1.ClusterRoleBinding{}
			err = json.Unmarshal(yToJ, &rb)
			if err != nil {
				return err
			}
			katesInput.Data = rb
			katesOutput, err = kates.RoleBindings(&katesInput)
		case K8KindSecret:
			svc := &apiv1.Secret{}
			err = json.Unmarshal(yToJ, &svc)
			if err != nil {
				return err
			}
			katesInput.Data = svc
			katesOutput, err = kates.Secret(&katesInput)
		case K8KindService:
			svc := &apiv1.Service{}
			err = json.Unmarshal(yToJ, &svc)
			if err != nil {
				return err
			}
			katesInput.Data = svc
			katesOutput, err = kates.Service(&katesInput)
		case K8KindServiceAccount:
			sc := &apiv1.ServiceAccount{}
			err = json.Unmarshal(yToJ, &sc)
			if err != nil {
				return err
			}
			katesInput.Data = sc
			katesOutput, err = kates.ServiceAccount(&katesInput)
		case K8KindStatefulSet:
			ss := &appsv1.StatefulSet{}
			err = json.Unmarshal(yToJ, &ss)
			if err != nil {
				return err
			}
			katesInput.Data = ss
			katesOutput, err = kates.StatefulSet(&katesInput)
		case K8KindStorageClass:
			sc := &storagev1.StorageClass{}
			err = json.Unmarshal(yToJ, &sc)
			if err != nil {
				return err
			}
			katesInput.Data = sc
			retry = false
			katesOutput, err = kates.StorageClass(&katesInput)
		default:
			err = errors.New(fmt.Sprintf("No matching kind: %s", m["kind"]))
		}

		// See if it failed, retries for 30 seconds to make sure no errors due to uncreated namespaces
		if err != nil {
			if retries == 30 && retry {
				return err
			} else if retries <= 30 && retry {
				log.Println("Retrying: ", m["kind"])
				time.Sleep(1000 * time.Millisecond)
			}
		} else {
			retries = 31
		}
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
					err = k.postTie(subSubDirFileInfo, subSubFilePath)
					if err != nil {
						workDirErrCh <- 1
					}
				}
			}
		}
		_ = subSubDir.Close()
	}
}*/
