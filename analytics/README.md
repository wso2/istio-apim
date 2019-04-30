## Developer Guide

This guide is for compiling the source code for analytics.

##### Prerequisites

- Maven
- Docker

##### Compile source code and build the docker image

```
mvn clean install
```

**Note:**

When you build the source code,

1. Creates the docker image
2. Copy jar files to install/analytics/resources/lib location

