# PersistentVolumeKubeController

This is a custom controller for Kubernetes. It watches for PersistentVolumes (PV)  objects and manages them.

When a PV is created, the controller will create an event of creation. By default, kubernetes cluster 1.29 does not create PV creation event. And it is useful to register this types of events for PVC custom storage class !

## Running controller

```bash
cd pvController
go get <dependencies>
go build -o pv_controller .
./pv_controller
```
