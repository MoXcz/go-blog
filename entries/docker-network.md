---
date: 30-Jun-2025
title: Docker network
---

## What is a container

Containers offer an abstraction that facilitates isolation between processes:

- *Chroot*. Short for *change root*, `chroot` can be used to *jail* processes by isolating them into their own *root*: isolates files, but processes are still visible.
- *Namespaces*. Isolates the process itself to *limit their access* without the process being aware of this limitation: isolates processes, but resources are still shared.
- *Cgroups*. Isolate environment resources and avoid memory leak problems: isolate resources.

The purpose of a container is to be able to run applications without having to simulate a *hardware interface* to do so, that is, without needing *another computer* (or abstraction of a computer under a virtual machine). Which is why this *isolation* is necessary, as everything is performed under the same computer/device.

## What is Docker

Docker is used for *containerization* of applications, which are like lightweight virtual machines *with batteries included* (libraries, executables, and any other dependencies may be needed).

{{< image src="/img/docker.svg" alt="docker-infrastructure" position="center" >}}

Just to ease in how Docker works, there are 3 main components of Docker: the **daemon** (the actual program), the **GUI** (graphical interface), and the **CLI** (command line interface).

The instructions below are provided using the CLI, but they should be easy to replicate using the GUI if that's preferred.

### Installation

Installing Docker is fairly simple, just follow the [Docker documentation](https://docs.docker.com/engine/install/#supported-platforms). In my case that I'm on Debian I could do this:

```sh
curl -O https://desktop.docker.com/linux/main/amd64/docker-desktop-amd64.deb
sudo apt install ./docker-desktop-amd64.deb

# And then just enable/start the daemon
systemctl --user enable docker-desktop
systemctl --user start docker-desktop
# Verify that it was installed correctly
docker version
```

## Docker applications

But how are these containers created? There are some *images* which provide instructions in how to *build* the container so that it can be started:

```sh
# this will download the image (if it's not installed)
# provided by nginx and start the container of that image
docker run -it --rm --name nginx -p 8080:8080 nginx:latest
```

Now, if you open your browser in [localhost:8080](http://localhost:8080) you should be able to see the welcome page provided by the nginx web server.

This is where one might start to see how *awesome* containers can be. Nginx could also have been installed locally, but that would require not only some extra steps but probably some extra knowledge in the tooling (which in this case is not as hard to install/configure, but I digress).


## Docker networks

The *network* in this context refers to how to use these *containers* inside a network from within the container abstraction and outside it from the host machine (as seen with the nginx web server).

Fortunately, Docker already has some pretty neat networks by default: `bridge`, `host`, and `none` that can be used to explain how networks are managed inside containers.

```sh
docker network ls # list the networks
```

When first installing Docker an interface called `docker0` is created, which is used as the default `bridge` network as mentioned above.

```sh
netstat -i # see the interfaces in your system
ip a       # should also work
```

In the following examples of how these networks work I will be using an image that I *constructed* by providing the following instructions inside a file called `Dockerfile`:

```dockerfile
FROM golang:1.24-alpine
WORKDIR /app
COPY . .

RUN go build -o server .
RUN apk add --no-cache curl

EXPOSE 8080
CMD ["./server", "-addr=:8080"]
```

This will just build and start a Go server. You can also use this application by *cloning* the repository as seen below:

```sh
git clone https://github.com/MoXcz/docker-network-practice
cd docker-network-practice

docker build . -t srv-test
docker run -it --rm --name srv -p 8080:8080 srv-test
# open a new terminal and issue a GET request
curl localhost:8080
```

Or, just visit [localhost:8080](http://localhost:8080) inside a browser just like with nginx.

The *response* (either in the browser or the terminal) should be something like `Hi from :8080 (hostname: da12f89eac4d IP: [172.17.0.2])`.

The number after `hostname:` is the *container id*, which you can see using `docker ps` while the container is still running. The IP is assigned by Docker using the *default `bridge` network*.

### But, what is a `bridge`?

{{< image src="/img/bridge-networks.svg" alt="bridge-networks" position="center" >}}

A *bridge* is used as a literal bridge between different device (or groups of devices) found inside a network.

In the case of Docker the bridge is between the containers and the host:

{{< image src="/img/docker-bridge.svg" alt="docker-bridge" position="center" >}}

> Note: To reset the state of the containers just stop (`docker stop <container_id>` and start them again `docker run`)

## `bridge`

So, this means that the `bridge` exists within its own private network that is then accessed through the host by *mapping* the ports (done through the `-p` option).

```sh
docker run -d --rm --name srv-1 -p 8080:8080 srv-test
docker run -d --rm --name srv-2 -p 8081:8080 srv-test
```

This will initialize two containers in the `bridge` network. This means that *both containers are in the same network*, which should made them visible to one another:

```
docker exec -it srv-1 sh
/app # curl localhost:8080
Hi from :8080 (hostname: e4067bc988c3 IP: [172.17.0.2])
/app # curl localhost:8081
curl: (7) Failed to connect to localhost port 8081 after 0 ms: Could not connect to server
```

What happened? `localhost` references the host, which when inside a container *the container is the host*, so `srv-1` does not know anything about `srv-2` *inside itself*, it knows there's a network (`bridge`).

So what can we do? Be more *specific* and get access through the network:

```sh
docker network inspect bridge # get IPs of containers
```

By using the IP of the container directly (remember that both are on the same `bridge` network) `srv-1` can make requests to `srv-2` (in this case its IP is `172.17.0.3`):

```
docker exec -it srv-1 sh
/app # curl localhost:8080
Hi from :8080 (hostname: e4067bc988c3 IP: [172.17.0.2])
/app # curl 172.17.0.3:8080
Hi from :8080 (hostname: 6f8382a402b7 IP: [172.17.0.3])
```

### Container IPs?

IP, the Internet Protocol, is used as a sort of *address* to recognize different devices. In the case of Docker, it creates a subnet in `172.17.0.0/16`, which is just a way of saying that the IPs (addresses) are delimited by this *subnet*.

Which is why the containers above return `172.17.0.2` and `172.17.0.3`.

Also, similar to the previous containers, these two can also be accessed through the host:

```sh
curl localhost:8080
curl localhost:8081
```

But what if we don't want the containers to be accesible from the host? Enter the `none` network.

## `none`

When creating a container it's always assigned to `bridge` even if it's not *mapped* to the host:

```
docker run -d --rm --name srv-3 srv-test
docker exec -it srv-1 sh
curl 172.17.0.4:8080
Hi from :8080 (hostname: 52225705d65a IP: [172.17.0.4])
```

But if we try to access the container from outside the network in the host there are no ways to access

