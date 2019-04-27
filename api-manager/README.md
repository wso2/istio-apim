## Developer Guide

This guide is for compiling the source code for api manager custom executor.

##### Prerequisites

- Maven


##### Compile source code

```
mvn clean install
```

##### Copy generated jars to install location

```
cp ./distribution/target/files/lib/* ./../install/api-manager/dropins/
cp ./distribution/target/files/non-osgi/* ./../install/api-manager/lib/
```