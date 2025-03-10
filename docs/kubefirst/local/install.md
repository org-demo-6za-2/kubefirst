# Local Platform Installation

`kubefirst` is the name of our command line tool that installs the Kubefirst platform to your local or cloud environment.

To use the local version of Kubefirst, you will need to have [Docker installed](https://docs.docker.com/get-docker/). You will also need a GitHub account: GitLab for local, and local git repositories are not supported yet.

![Kubefirst local installation diagram](../../img/kubefirst/local/kubefirst-cluster-create.png)

### Prerequisites

- [To install kubefirst CLI](../overview.md#how-to-install-kubefirst-cli)
- [To install docker](https://docs.docker.com/get-docker/)
- [A personal github account](https://github.com/) (`gitops` and `metaphor-frontend` repositories will be created in your account and should not preexist)

## 2 Hour Expiration Warning

The ngrok tunnel used for kubefirst local has a 2-hour expiration unless you create an account with ngrok. This expiration will prevent you from using automated infrastructure as code through atlantis, but the rest of the platform will continue to function beyond that ngrok evaluation period. [Create an account with ngrok](https://dashboard.ngrok.com/signup) to prevent this.

## Create your new local cluster

To create a new Kubefirst cluster locally, run

```shell
kubefirst local
```

More information on `kubefirst local`, including optional flags, can be found [in the CLI Documentation](../../tooling/kubefirst-cli.md)

If your run is not successful, errors and troubleshooting information will be stored in a local log file specified during the installation run.

This will be followed by the instructions prompt to populate the `KUBEFIRST_GITHUB_AUTH_TOKEN` env variable for your [github token](../../explore/github-token.md). Press `ENTER` and follow the prompt to continue.

Please export a `KUBEFIRST_GITHUB_AUTH_TOKEN` if you need your ephemeral environment for more than 8 hours. The ephemeral GitHub tokens that we can create for you expire after 8 hours.

The installation process may take a few minutes. If you are successful you should see:

```shell
Cluster "kubefirst" is up and running!
```

#### Installed Applications

Kubefirst implicitly contains many applications to provide starting capabilities for new users. Operational knowledge of all applications is not necessary to begin using Kubefirst, but is useful to understand your cluster.

A newly created local Kubefirst cluster contains:

- A private repo named `gitops`. The applications that you build and release on the kubefirst platform will also be registered here in the development, staging, and production folders. 
- [Argo CD](https://github.com/argoproj/argo-cd) - GitOps Continuous Delivery
- [Argo Workflows](https://argoproj.github.io/argo-workflows/) - Application Continuous Integration
- [Atlantis](https://www.runatlantis.io/) - Terraform Workflow Automation
- [Chart Museum](https://github.com/helm/chartmuseum) - Helm Chart Registry
- [External Secrets](https://github.com/external-secrets/kubernetes-external-secrets) - Syncs Kubernetes secrets with Vault secrets
- [GitHub Action Runner](https://github.com/features/actions) - Self Hosted GitHub Action Runner
- [Metaphor](https://github.com/kubefirst/metaphor-frontend-template) - A sample app to demonstrate CI/CD in on Kubernetes. Contains Devlopment, Staging, and Production environments.
- [Traefik](https://github.com/traefik/traefik) - Default Ingress Controller for K3D Clusters
- [Vault](https://github.com/hashicorp/vault) - Secrets Management

### How to resolve HTTPS Certificate Warnings

To resolve the warning that the browser shows when you access one of your applications, run the command:
```shel
mkcert -install
```
 We use [Mkcert](https://github.com/FiloSottile/mkcert) to generate local certificates and serve `https` with the Traefik Ingress Controller.

During installation, Kubefirst generates these certificates and pushes them to Kubernetes as secrets to attach to Ingress resources. The browser does not recognize auto-assigned certificates as trusted certificates and will generate security errors. 

This step will install the CA (Certificate Authority) of MkCert in your trusted store and will allow the browser to trust in certificates generated by your Kubefirst local install.

### Atlantis and Ngrok integration

[Ngrok](https://ngrok.com/) is a tool that allows Kubefirst to expose a local server to the internet via an [ngrok Secure Tunnel](https://ngrok.com/docs/secure-tunnels/). Kubefirst opens an ngrok Secure Tunnel tunnel during the installation to send events to Atlantis. When the installation finishes, the terminal window hangs at the handoff screen.
If the handoff screen in your terminal is closed, the Kubefirst installation terminates and the Ngrok Secure Tunnel is closed.

During cluster provisioning, Terraform communicates with the host machine to create the desired resources. When Atlantis is installed via Kubefirst, it will use ngrok to expose the Atlantis server to the internet via [webhook](https://zapier.com/blog/what-are-webhooks/?utm_source=google&utm_medium=cpc&utm_campaign=gaw-usa-nua-search-blog-dsa&utm_adgroup=DSA-Guides-What_are_webhooks&utm_term=&utm_content=_pcrid_630760751271_pkw__pmt__pdv_c_slid__pgrid_145358980000_ptaid_dsa-1873981911115_&gclid=Cj0KCQiAw8OeBhCeARIsAGxWtUxZLa8mXxQUt484tVLVjTCCl3zlHEmklG2Gu-EXdy1u521wyIg6EcoaAlS5EALw_wcB).

## After installation

After the ~5 minutes installation, your browser will launch a new tab to the [Kubefirst Console application](https://github.com/kubefirst/console), which will help you navigate your new suite of tools running in your local k3d cluster.

Continue your journey: 

- [Explore your installation](./explore/overview.md)
- [Destroying](./destroy.md)
