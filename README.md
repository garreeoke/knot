# Knot - Deploy repo of kubernetes spec files to a cluster

Deploy Kubernetes spec files from a local directory or github repo. Originally created for use with 
VMware Enterprise PKS to be used in Post-Deployment box.  However, I quickly found other uses for it.
* Use with pipelines or other utilities to easily deploy from a specified location
* Setup a kubernetes CronJob to deploy from a directoy at a specified interval.  Keeps everything updated.


## Easiest Ways to Run ##
* [Docker run](https://github.com/garreeoke/knot#docker-example)
* [Kubernetes job](https://github.com/garreeoke/knot#kubernetes)

## Example repos to pull from
* [My repo](https://github.com/garreeoke/k8_setup)

## ENV Variables to set
* KNOT_TYPE - Where to get the files.
    * Supported values: [github, local]
        * If local, must mount desired directory to /knot/files
* KNOT_URI - Path to get the values if github
    * I.E. - owner/garreeoke/repository/k8_setup/branch/master
* KNOT_AUTH - How to authenticate
    * local - Will try to read kubeconfig in users home directory
    * cluster - Use this when running as a job on K8s cluster
* KNOT_ACTION - What action to take
    * Supported: [create, update, dynamic]
        * dynamic will determine if a create or update is needed, useful for running CronJob
* KNOT_WHITELIST - Comma separated list of sub-directories to use. Not specified means deploy everything.
* GITHUB_USER - Specify github user if repo is not public
* GITHUB_TOKEN - Specify github access token for user if repo is not public

## Kubernetes
Look in the k8s directory for examples of configuring a job to run.
* With ConfigMap
    * \# kubectl create -f knot_configMap.yaml
    * \# kubectl create -f knot_job_with_configMap.yaml
* Without ConfigMap
    * \# kubectl create -f knot_job_with_env.yaml
* As a CronJob
    * \# kubectl create -f knot_cronjob_with_env.yaml
    
After the job runs, the pod will still be available to look at the logs.  Once you are done looking at the logs,
delete the job.
    
## VMware PKS Enterprise
To use with PKS Enterprise, copy the contents of k8s/knot_pks.yaml to the desired plan in OpsMgr.
There is a box underneath "(Optional) Add-ons - Use with caution" in Plan configuration, paste contents of desired yaml there.
Be sure to modify the knot_pks.yaml (KNOT_URI environment variable) to point to the correct github repo.

## Docker Example

* Github
    * docker run -e "KNOT_AUTH=local" -e "KNOT_TYPE=github" -e "KNOT_ACTION=dynamic" -e "KNOT_URI=owner/garreeoke/repository/k8_setup/branch/master" -v /root/.kube/config:/root/.kube/config garreeoke/knot
* Local Directory
    * docker run -e "KNOT_AUTH=local" -e "KNOT_TYPE=local" -e "KNOT_ACTION=dynamic" -e "KNOT_WHITELIST=acme-air,d1" -v /root/.kube/config:/root/.kube/config -v /home/aaron/knot_test:/knot/files garreeoke/knot

## Notes
* Only single type yaml files are supported.
* The repo can have one level of sub-directories.  This is useful if you want to place all related yamls for
individual services in their own place.

