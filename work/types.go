package work

import (
	"github.com/garreeoke/kates"
	"k8s.io/client-go/kubernetes"
)

const FileDir = "/knot/files"

// Knot Type Constants
const (
	TypeGitHub = "github"
	TypeLocal = "local"
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
	K8KindNamespace 			= "Namespace"
	K8KindNetworkPolicy			= "NetworkPolicy"
	K8KindSecret				= "Secret"
)

// Auth contants
const (
	OnCluster = "cluster"
	Local     = "local"
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
	Auth           string `json:"auth,omitempty"`
	KubeConfigPath string `json:"kube_config_path,omitempty"`
	Operation      string `json:"operation,omitempty"`
	WorkDir        string `json:"work_dir,omitempty"`
	// Whitelist - List of sub-directories to try
	WhiteList	   []string `json:"white_list,omitempty"`
	Client         *kubernetes.Clientset
	Output         []*kates.Output `json:"output,omitempty"`
}

// Github info to get gitHub data
type GitHub struct {
	Path    string
	Token   string
	User    string
	WorkDir string
}
