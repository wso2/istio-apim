## Developer Guide

This guide is for compiling the source code for analytics.

##### Prerequisites

- Maven


##### Compile source code

```
mvn clean install
```

##### Copy generated jars to install location

```
cp ./distribution/target/files/lib/* ./../install/analytics/lib/
```