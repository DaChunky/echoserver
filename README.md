# Echoserver
Simple TCP echoserver to test TCP routings and settings in i.e. a K8S cluster or a container enviroment

# Build
Just clone the branch and run 
~~~~
make
~~~~
in the docker folder. It will build the server application and create a container image. 
## App
If you only want build the app, run
~~~~
make app
~~~~
## Container
If you only want to build the container, run
~~~~
make container
~~~~
# Usage
Per default the server app is listening on port 12345. If you want to run the server app on another port, at an argument
~~~~
echoserver 43251
~~~~
## Container
To run the container, you could use the image from Docker hub ([fschunke/echoserver](https://hub.docker.com/repository/docker/fschunke/echoserver)), which is build with an ARM7 processor.
~~~~
docker run -itd -e SEVER_PORT=42235 -p 43251:42235 fschunke/echoserver
~~~~
## Kubernetes
In the k8s folder you could find a deployment yaml and a configmap, which could be used to deploy the server app on a kubernetes cluster. The configmap is only working if you add an ingress controller to your cluster (i.e. [https://www.nginx.com/products/nginx-ingress-controller/](https://www.nginx.com/products/nginx-ingress-controller/)). Afterwards you have to edit the ingress controller deployment
~~~~
KUBE_EDITOR=nano kubectl edit deployment -n ingress-nginx -o yaml
~~~~
Add following lines to the args section of the container section
~~~~
...
  - args:
  ...
  - --tcp-services-configmap=$(POD_NAMESPACE)/tcp-services
  - --udp-services-configmap=$(POD_NAMESPACE)/udp-services
...
~~~~
The name `tcp-services` is the same as in the configmap file.
