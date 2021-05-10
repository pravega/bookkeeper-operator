# Configuring Bookkeeper Headless Service Name

By default bookkeeper headless service name is configured as  `[CLUSTER_NAME]-bookie-headless`.

```
bookkeeper-bookie-headless              ClusterIP   None             <none>        3181/TCP       4d15h
```
But we can configure the headless service name as follows:

```
helm install bookkeeper pravega/bookkeeper --set headlessSvcNameSuffix="headless"
```

After installation services can be listed using `kubectl get svc` command.


```
bookkeeper-headless         ClusterIP   None             <none>        3181/TCP       4d15h
```
