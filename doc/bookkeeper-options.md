## BookKeeper options

BookKeeper has many configuration options. The available options can be found [here](https://bookkeeper.apache.org/docs/4.7.0/reference/config/) and are expressed through the `options` part of the resource specification.

All values must be expressed as Strings.

Take metrics for example, here we choose codahale as our metrics provider. The default is Prometheus.

```
...
spec:
  options:
    enableStatistics: "true"
    statsProviderClass: "org.apache.bookkeeper.stats.codahale.CodahaleMetricsProvider"
    codahaleStatsGraphiteEndpoint: "graphite.example.com:2003"
    codahaleStatsOutputFrequencySeconds: "30"
...
```
### BookKeeper JVM Options

It is also possible to tune the BookKeeper JVM by passing customized JVM options. BookKeeper JVM Options
are for Bookkeeper JVM whereas the aforementioned BookKeeper options are for BookKeeper server configuration.

The format is as follows:
```
...
spec:
  jvmOptions:
    memoryOpts: ["-Xms2g", "-XX:MaxDirectMemorySize=2g"]
    gcOpts: ["-XX:MaxGCPauseMillis=20"]
    gcLoggingOpts: ["-XX:NumberOfGCLogFiles=10"]
    extraOpts: []
...
```
The reason that we are using such detailed names like `memoryOpts` is because the BookKeeper official [scripts](https://github.com/apache/bookkeeper/blob/master/bin/common.sh#L118) are using those and we need to override it using the same name. JVM options that don't belong to the earlier 3 categories can be mentioned under `extraOpts`.

There are a bunch of default options in the BookKeeper operator code that is good for general deployment. It is possible to override those default values by just passing the customized options. For example, the default option `"-XX:MaxDirectMemorySize=1g"` can be overridden by passing `"-XX:MaxDirectMemorySize=2g"` to
the BookKeeper operator. The operator will detect `MaxDirectMemorySize` and override its default value if it exists. Check [here](https://www.oracle.com/technetwork/java/javase/tech/vmoptions-jsp-140102.html) for more JVM options.

Default memoryOpts:
```
"-Xms1g",
"-XX:MaxDirectMemorySize=1g",
"-XX:+ExitOnOutOfMemoryError",
"-XX:+CrashOnOutOfMemoryError",
"-XX:+HeapDumpOnOutOfMemoryError",
"-XX:HeapDumpPath=" + heapDumpDir,
```
if BookKeeper version is greater or equal to 0.4, then the followings are also added to the default memoryOpts
```
"-XX:+UnlockExperimentalVMOptions",
"-XX:+UseCGroupMemoryLimitForHeap",
"-XX:MaxRAMFraction=2"
```

Default gcOpts:
```
"-XX:+UseG1GC",
"-XX:MaxGCPauseMillis=10",
"-XX:+ParallelRefProcEnabled",
"-XX:+AggressiveOpts",
"-XX:+DoEscapeAnalysis",
"-XX:ParallelGCThreads=32",
"-XX:ConcGCThreads=32",
"-XX:G1NewSizePercent=50",
"-XX:+DisableExplicitGC",
"-XX:-ResizePLAB",
```

Default gcLoggingOpts:
```
"-XX:+PrintGCDetails",
"-XX:+PrintGCDateStamps",
"-XX:+PrintGCApplicationStoppedTime",
"-XX:+UseGCLogFileRotation",
"-XX:NumberOfGCLogFiles=5",
"-XX:GCLogFileSize=64m",
```

### BookKeeper Custom Configuration

It is possible to add additional parameters into the BookKeeper container by allowing users to create a custom ConfigMap  and specify its name within the field `envVars` of the BookKeeper Spec. The following values need to be provided within this ConfigMap if we expect the BookKeeper cluster to work with Pravega.

| KEY | VALUE |
|---|---|
| *PRAVEGA_CLUSTER_NAME* | Name of Pravega Cluster using this BookKeeper Cluster |
| *WAIT_FOR* | Zookeeper URL |

The user however needs to ensure that the following keys which are present in BookKeeper ConfigMap which is created by the BookKeeper Operator should not be a part of this custom ConfigMap.

```
- BOOKIE_MEM_OPTS
- BOOKIE_GC_OPTS
- BOOKIE_GC_LOGGING_OPTS
- BOOKIE_EXTRA_OPTS
- ZK_URL
- BK_useHostNameAsBookieID
- BK_useHostNameAsBookieID
- BK_AUTORECOVERY
- BK_lostBookieRecoveryDelay
```
