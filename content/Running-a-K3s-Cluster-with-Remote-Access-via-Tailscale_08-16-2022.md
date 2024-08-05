## Running a K3s Cluster w/ Remote Access via Tailscale
In this tutorial I will explain how to get a K3s cluster running and how to access that cluster remotely using Tailscale. If you've found this post, it's likely safe to assume that you were looking for it and you can skip the next section. 

### Why would you want to do this?
The primary reason I've decided to do this is learning. Container deployments are standard at work, but the management and set up is handled by other teams. It's pretty easy to set up a cluster for learning with a cloud provider or even minikube. I went the raspberry pi route because I like hardware and its nice to be able to point to something physical. If you give me a reason to buy a computer, I probably will. 

I'm using Tailscale because sometimes I'm not at home, but I still want to play with my stuff. 

### Setting up K3s
First, we will install the 64bit image of raspbian lite on each of the raspberry pis. The raspberry pi imager is a great tool for this job and these days it will allow you to enable ssh, setup a wifi connection, and change the username and hostname. This makes life really easy. All of this stuff could be done relatively easily before and if you're curious you can still learn to do it all yourself. Anyway, I generally choose to have my pis setup so that each has the username `worker` and the hostname as `pi{# in cluster}`. For example: `worker@pi0, worker@pi1`. 

Once that is done, you should run the following command for each pi `ssh-copy-id worker@pi0`. This step is required in order to use `k3sup`, which is a nice and easy way to install K3s.

Details for installing `k3sup` can be found on [its GitHub page](https://github.com/alexellis/k3sup). 

### Install the server
```bash
k3sup install --ip {ip of server} --user {username... likely "worker" or "master"}
```

### Add the workers to the cluster
```bash
k3sup join --ip {ip of worker} --server-ip {ip of server} --user {username of worker}
```

You can verify that the above two steps completed as expected by running the following two commands: 

```bash
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node -o wide
```

If the pis aren't all there, then something has gone wrong. 

### Setting up Tailscale
First of all, you should make an account with Tailscale. It's free! On each raspberry pi run the following two commands: 

```bash
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up
```

At this point you have a working K3s cluster and the ability to ssh into each node using its Tailscale IP address from any device that is connected to your Tailscale account.

However, if you try to run a kubectl command the cluster's local network it won't work. To get that working you can simply edit the `kubeconfig` file that k3sup created earlier. You can open it like this `vim ~/kubeconfig` and then simply change the IP address in the server line. Instead of using the local IP address, use the Tailscale IP address. However, do not remove the `https://` or the port. 

All done. If things don't work, you've likely goofed. 

Also, be warned that messing with dns stuff in Tailscale can end badly. If you override the local dns because you want to deploy a pi-hole instance, things will break. I'd recommend you just set the dns on the devices that you want the pi-hole to filter on (probably just your laptop and phone). 



