# Knot - Deploy repo of kubernetes spec files to a cluster

Each time a kubernetes cluster is created, there is probably a set of yaml files that need to be deployed.  Use Knot to
read from a repo (github) and deploy all the desired yaml.  Knot was originally intended for use while deploying kubernetes
clusters using VMware Enterprise PKS.  However, it can easily be used with any kubernetes cluster.

## ENV Variables to set
* KNOT_TYPE - Where to get the files.
    * Supported values: [github]
* KNOT_URI - Path to get the values
    * I.E. - owner/garreeoke/repository/k8_setup/branch/master
* KNOT_AUTH - How to authenticate
    * local - Will try to read kubeconfig in users home directory
    * cluster - Use this when running as a job on K8s cluster
* GITHUB_USER - Specify github user if repo is not public
* GITHUB_TOKEN - Specify github access token for user if repo is not public

## Kubernetes
Look in the k8s directory for examples of configuring a job to run.  Examples below is the yaml files in the k8s folder.
* With config map
    * kubectl create -f knot_configMap.yaml
    * kubectl create -f knot_job_with_configMap.yaml
* Without configMap
    * kubectl create -f knot_job_with_env.yaml
    
## VMware PKS Enterprise
To use with PKS Enterprise, copy the contents of k8s/knot_pks.yaml to the desired plan in OpsMgr.
There is a box underneath "(Optional) Add-ons - Use with caution" in Plan configuration, paste contents of desired yaml there.
Be sure to modify the yaml file to point to the correct github repo.

* \- name: KNOT_URI
    value: "owner/garreeoke/repository/k8_setup/branch/master"

## Docker Example

* docker run -e "KNOT_AUTH=local" -e "KNOT_TYPE=github" -e "KNOT_URI=owner/garreeoke/repository/k8_setup/branch/master" -v /Users/torgersona/.kube/config:/root/.kube/config garreeoke/knot

## Notes
* Only single entity yaml files are supported right now.  For example, if you have something to deploy that is a
service and a deployment.  A K8s yaml file for each one should be used.
