# K8 Cluster Setup

The Cluster Setup supports the following providers:

### Repos used to make this:

### Cluster Builder runs via environment variables:

Cluster Builder get's all of it's settings from environment variables. We also use environment variables for high-level control plane. 

 * CLOUD_PROVIDER
    * This must be [ gke | aws ]

 * CLOUD_ACTION
    * This must be [ create | (delete | destroy) | upgrade ]

