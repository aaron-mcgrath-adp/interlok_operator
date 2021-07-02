# GO environment set-up for Windows 10

On your windows host;

 - Install WSL 2 (Ubuntu)
 - Install Minikube

## Install GO

Download the latest package - https://golang.org/dl/go1.16.5.linux-amd64.tar.gz

Change directory to the download location and then remove any previous install and unpack the TAR like this;
```
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.16.5.linux-amd64.tar.gz
```

Then modify the PATH to include the GO binaries.  Simply add the following line to your $HOME/.profile;

```
export PATH=$PATH:/usr/local/go/bin
```

You may need to restart your linux subsystem for the changes to take effect.

You should then be able to simply check the installation on the command line like so;

```
go version
```

## Install MAKE into Ubuntu

```
sudo apt-get build-essential
```

MAKE should now report it's version like this;
```
make --version
```

## Install Minikube

You can either install Minikube directly on top of your WSL (inside your ubuntu instance) or if you already have minikube installed (and running) on your windows host then you can simply create a new file in the following directory;
```
~/.kube/config
```

And the contents of that file simply points your __kubectl__ install (next section) to connect to your windows hosts minikube.  Obviously check for correct paths, but the contents will be;
```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /mnt/c/Users/Aaron/.minikube/ca.crt
    extensions:
    - extension:
        last-update: Mon, 14 Jun 2021 13:45:03 BST
        provider: minikube.sigs.k8s.io
        version: v1.21.0
      name: cluster_info
    server: https://127.0.0.1:50445
  name: minikube
contexts:
- context:
    cluster: minikube
    extensions:
    - extension:
        last-update: Mon, 14 Jun 2021 13:45:03 BST
        provider: minikube.sigs.k8s.io
        version: v1.21.0
      name: context_info
    namespace: default
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate: /mnt/c/Users/Aaron/.minikube/profiles/minikube/client.crt
    client-key: /mnt/c/Users/Aaron/.minikube/profiles/minikube/client.key

```

The next section will describe installing kubectl which will connect to your hosts version of Minikube.  If it fails to connect (see next section), then you can just copy your config file from your host (C:\Users\user\.kube\config) into your ubuntu home directory as shown above and simply change the paths.
Do note the paths need to start with __/mnt/c/__, this is the subsystems way to get access to the hosts C drive root.

## Install Kubectl

Update the apt package index and install packages needed to use the Kubernetes apt repository:

```
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl
```

Download the Google Cloud public signing key:

```
sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
```

Add the Kubernetes apt repository:

```
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
```

Update apt package index with the new repository and install kubectl:

```
sudo apt-get update
sudo apt-get install -y kubectl
```

Now we can test that not only kubectl has installed correctly, but also connects to your windows hosts Minikube instance;
```
kubectl version
```

You're looking to get back two lines, the first for the client and the second for the server.  If it cannot connect to your server your second line in the output will show an error;
```
aaron@DESKTOP-EDI1AH1:~$ kubectl version
Client Version: version.Info{Major:"1", Minor:"21", GitVersion:"v1.21.1", GitCommit:"5e58841cce77d4bc13713ad2b91fa0d961e69192", GitTreeState:"clean", BuildDate:"2021-05-12T14:18:45Z", GoVersion:"go1.16.4", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"21", GitVersion:"v1.21.1", GitCommit:"132a687512d7fb058d0f5890f07d4121b3f0a2e2", GitTreeState:"clean", BuildDate:"2021-05-12T12:32:49Z", GoVersion:"go1.16.4", Compiler:"gc", Platform:"linux/amd64"}
```

## Install the Kubernetes Operator SDK

```
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
```

Download the binary for your platform:

```
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.8.0
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
```

Verify the downloaded binary:

```
gpg --keyserver keyserver.ubuntu.com --recv-keys 052996E2A20B5C7E
```

Download the checksums file and its signature, then verify the signature:

```
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt.asc
gpg -u "Operator SDK (release) <cncf-operator-sdk@cncf.io>" --verify checksums.txt.asc
```

You should see something similar to the following:

```
gpg: assuming signed data in 'checksums.txt'
gpg: Signature made Fri 30 Oct 2020 12:15:15 PM PDT
gpg:                using RSA key ADE83605E945FA5A1BD8639C59E5B47624962185
gpg: Good signature from "Operator SDK (release) <cncf-operator-sdk@cncf.io>" [ultimate]
```

Make sure the checksums match:

```
grep operator-sdk_${OS}_${ARCH} checksums.txt | sha256sum -c -
```

You should see something similar to the following:

```
operator-sdk_linux_amd64: OK
```

Install the release binary in your PATH:

```
chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
```

## Notes on developer usage

### MAKE

When you modify any of the __*_types.go__ you need to regenerate the source code, you do that by changing directory to the root of the GO project and executing;
If you don't want to execute kubectl to install your updated components into minikube with MAKE, then remove the __make install__ from the below.
```
make generate; make manifests; make install
```

Essentially if you run the __install__ command you are calling a __kubectl apply -f __ to your newly generated CRD's.

Which then allows you to query for such things in Minikube with a no resources found, rather than an error telling you the resource doesn't exist;
```
$kubectl get Interlok
No resources found in default namespace.
```

If you apply the sample Interlok instance found in __./config/samples__;

```
kubectl apply -f ./config/samples/blah.yaml
```

Then you can also do a __make run__ on the command line to test your operator.

## Docker Image

The latest version can be found here; https://hub.docker.com/repository/docker/aaronmcgrathadpx/interlok-operator

## Deployment

You can deploy the operator in one of two ways.  The first requires the development environment as discussed above and then you can simply execute one of the following on your linux command line at the root of the interlok-operator;

```
$ make install
$ make run
```

Make install will perform apply the CRD.

Make run, will simply execute the main.go class which will start the operator in your cluster.

The second option is simply to apply the yaml files in the config directory.  You'll find everything from the CRD, RBAC permissions, controller manager through to a sample Interlok deployment for our controller to manage. 
