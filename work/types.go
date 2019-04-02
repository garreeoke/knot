package work

import (
	"github.com/garreeoke/kates"
	"k8s.io/client-go/kubernetes"
)

const TmpDir = "/knot/files"

// Knot Type Constants
const (
	TypeGitHub = "github"
)

// K8 constants
const (
	K8KindDeployment            = "Deployment"
	K8KindStatefulSet           = "StatefulSet"
	K8KindJob                   = "Job"
	K8KindCronJob               = "CronJob"
	K8KindConfigMap             = "ConfigMap"
	K8KindStorageClass          = "StorageClass"
	K8KindIngress               = "Ingress"
	K8KindService               = "Service"
	K8KindReplicationController = "ReplicationController"
	K8KindDaemonSet             = "DaemonSet"
	K8KindServiceAccount        = "ServiceAccount"
	K8KindClusterRole           = "ClusterRole"
	K8KindClusterRoleBinding    = "ClusterRoleBinding"
	K8KindRole                  = "Role"
	K8KindRoldBinding           = "RoleBinding"
)

// Auth contants
const (
	OnCluster = "cluster"
	Local = "local"
)

// Source is the interface to implement for different types of locations
type Source interface {
	GetFiles() (string, error)
	GetPath() string
	SetWorkDir(string)
	GetWorkDir() string
}

// Knot
type Knot struct {
	Auth    string `json:"auth,omitempty"`
	KubeConfigPath string `json:"kube_config_path,omitempty"`
	Action  string `json:"action,omitempty"`
	WorkDir string `json:"work_dir,omitempty"`
	Client  *kubernetes.Clientset
	Output  []*kates.Output `json:"output,omitempty"`
}

// Github info to get gitHub data
type GitHub struct {
	Path    string
	Token   string
	User    string
	WorkDir string
}
