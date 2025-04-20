# PersistentVolumeKubeController

This is a custom controller for Kubernetes. It watches for PersistentVolumeClaim (PVC) objects and manages PersistentVolumes (PV) for them.

When a PVC is created, the controller makes a PV for it. If the PVC is deleted, the controller deletes the PV. It also updates the PV if needed.

## How it works

The controller runs inside the cluster and uses the Kubernetes API to watch PVCs. It creates and manages PVs to match those claims.

Here is a diagram that shows the idea:

![controller-diagram](https://raw.githubusercontent.com/andresdossantosd/PersistentVolumeKubeController/main/controller-diagram.png)

- The controller is running as a pod in the cluster.
- It listens for changes in PVCs on ETCD.
- When a PVC is created, it makes a PV for it.
- If a PVC is updated or deleted, it updates or deletes the PV too.

## Run it

You can build the controller and run it inside a Kubernetes cluster.

```bash
go build -o controller .
```