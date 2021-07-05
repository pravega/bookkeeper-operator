# Troubleshooting

## Bookkeeper Cluster Issues

* [Certificate Error: Internal error occurred: failed calling webhook](#certificate-error-internal-error-occurred-failed-calling-webhook)
* [Invalid Cookie Exception](#invalid-cookie-exception)
* [Unrecognized VM option](#unrecognized-vm-option)

## Bookkeeper operator Issues
* [Operator pod in container creating state](#operator-pod-in-container-creating-state)

## Certificate Error: Internal error occurred: failed calling webhook

While installing bookkeeper, if we get the error as  below,
```
helm install bookkeeper charts/bookkeeper
Error: Post https://bookkeeper-webhook-svc.default.svc:443/validate-bookkeeper-pravega-io-v1alpha1-bookkeepercluster?timeout=30s: x509: certificate signed by unknown authority
```
We need to ensure that certificates are installed before installing the operator. Please refer [prerequisite](../charts/bookkeeper-operator/README.md#Prerequisites)

## Invalid Cookie Exception

While installing bookkeeper, if the pods are not coming to ready state `1/1` and in the bookie logs if the error messages are seen as below,

```
2020-06-26 09:03:34,893 - ERROR - [main:Main@223] - Failed to build bookie server
org.apache.bookkeeper.bookie.BookieException$InvalidCookieException:
        at org.apache.bookkeeper.bookie.Bookie.checkEnvironmentWithStorageExpansion(Bookie.java:470)
        at org.apache.bookkeeper.bookie.Bookie.checkEnvironment(Bookie.java:252)
        at org.apache.bookkeeper.bookie.Bookie.<init>(Bookie.java:691)
        at org.apache.bookkeeper.proto.BookieServer.newBookie(BookieServer.java:137)
        at org.apache.bookkeeper.proto.BookieServer.<init>(BookieServer.java:106)
        at org.apache.bookkeeper.server.service.BookieService.<init>(BookieService.java:43)
        at org.apache.bookkeeper.server.Main.buildBookieServer(Main.java:301)
        at org.apache.bookkeeper.server.Main.doMain(Main.java:221)
        at org.apache.bookkeeper.server.Main.main(Main.java:203)
```

we need to ensure that znode entries are cleaned up from previous installation. This can be done by either cleaning up znode entries from zookeeper nodes or by completely reinstalling zookeeper.

## Unrecognized VM option

While installing bookkeeper, if the pods don't come up to ready state and the logs contain the error shown below

```
Unrecognized VM option 'PrintGCDateStamps'
Error: Could not create the Java Virtual Machine.
Error: A fatal exception has occurred. Program will exit.
```
This is happening because some of default JVM options added by the operator are not supported by Java version used by bookkeeper. This issue can therefore be resolved by setting an additional JVM option `IgnoreUnrecognizedVMOptions` while installing the bookkeeper cluster as shown below.

```
helm install [RELEASE_NAME] pravega/bookkeeper --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set 'jvmOptions.extraOpts={-XX:+IgnoreUnrecognizedVMOptions}'
```

## Operator pod in container creating state

While installing operator, if the operator pod goes in `ContainerCreating` state for long time, make sure certificates are installed correctly. Please refer [prerequisite](../charts/bookkeeper-operator/README.md#Prerequisites)
