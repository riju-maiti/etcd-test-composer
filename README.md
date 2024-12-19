# Etcd-test-composer

## Purpose

The purpose of this repository is to serve as a reference point for using [Antithesis](https://antithesis.com/) tooling when testing on the [Antithesis platform](https://antithesis.com/product/what_is_antithesis/). This repository tests for violations of [etcd](https://etcd.io/)'s [linearizability](https://etcd.io/docs/v3.5/learning/api_guarantees/) guarantee.  

## Setup

There are four containers running in this system: three (`etcd0`, `etcd1`, `etcd2`) that make up an etcd cluster and one (`workload`) that ["makes the system go"](https://antithesis.com/docs/getting_started/basic_test_hookup/): in this case, by taking operations against etcd and confirming those operations were linearizable. When the workload container starts, the `entrypoint.py` script runs and confirms that all of the etcd hosts are available before [signaling the software is ready to test](https://antithesis.com/docs/getting_started/basic_test_hookup/#ready-signal). 

## Test Composer Drivers

Antithesis's [Test Composer](https://antithesis.com/docs/test_templates/) framework allows for modular test definition, that, when running on Antithesis, takes advantage of the Antithesis platform for things like parallelism, test length, and command order.

Executables become drivers based on [their absolute directory location and names](https://antithesis.com/docs/test_templates/first_test/#structuring-test-templates). In `test-template/Dockerfile`, you can see that two drivers get defined in the `workload` container image: `/opt/antithesis/test/v1/main/parallel_driver_generate_traffic.py` and `/opt/antithesis/test/v1/main/serial_driver_validate_operations`. *Drivers can be defined on any container in the system.* 

### Parallel Driver

`parallel_driver_generate_traffic.py` is a [parallel driver](https://antithesis.com/docs/test_templates/test_composer_reference/#parallel-driver) that submit operations against the etcd cluster. 0 to many of these drivers can ran at once. 

### Serial Driver

`serial_driver_validate_operations.go` is a [serial driver](https://antithesis.com/docs/test_templates/test_composer_reference/#serial-driver-command) that validates the operations that have happened against the etcd cluster. 

## SDK Usage

This repository includes the use of Antithesis's Python and Go SDKs. 

### setupComplete

The ["setupComplete"](https://antithesis.com/docs/generated/sdk/python/antithesis/lifecycle.html#setup_complete) signals that the system is ready to test: 

`setup_complete({"Message":"ETCD cluster is healthy"})`

### Assertions

[Antithesis SDKs allow users to define test properties](https://antithesis.com/docs/using_antithesis/sdk/#test-properties) directly within their application. There are two types of properties found in this repository. 

#### Sometimes Assertions

These [sometimes assertions](https://antithesis.com/docs/properties_assertions/properties/#sometimes-properties) confirm intended funcitonality happens- in this case, that operations happen against the etcd cluster and that validation occurs. For example, in `parallel_driver_generate_traffic.py`: 

`sometimes(success,"client can make successful put requests",None)`

#### Always Assertions

These [always assertions](https://antithesis.com/docs/properties_assertions/properties/#always-properties) confirm that guarantees are not violated- in this case, that means there are no operations that aren't serializable. For example, in `serial_driver_validate_operations.go`: 

`assert.Always(res == porcupine.Ok, "Operations against the cluster are linearizable", nil)`

### Randomness

Using randonmess in testing allows for varied execution of what's being tested. [The Antithesis SDK](https://antithesis.com/docs/using_antithesis/sdk/#randomness) provides an easy interface to get convenient and structured random values while also making it easier for the Antithesis Platform to learn what sorts of inputs to provide.

## Testing Locally

Before running your application on the Antithesis platform, a common approach to confirming your driver functionality works is by running it locally. 

While the SDK assertions won't be evaluated locally, you can still ensure everything is set up correctly. There are three steps to do this: 

1. Pull the bitnami/etcd:3.5 image using the following command: 

`docker pull bitnami/etcd:3.5`

2. Build the workload image. From within the `/test-template` directory, run the following command: 

`docker build . -t workload:latest`

3. run `docker-compose up` from the root directory to start all containers defined in `docker-compose.yml`

4. After the workload container has signaled `setupComplete` (or printed `cluster is healthy`), you can run the parallel driver 1 to many times via `docker exec`: 

`docker exec -it workload python3 /opt/antithesis/test/v1/main/parallel_driver_generate_traffic.py`

5. After that completes, you can run the serial driver in the same fashion: 

`docker exec -it workload /opt/antithesis/test/v1/main/serial_driver_validate_operations`

You should see a message: `antithesis-porcupine: Validate done`

You've now validated that your test is ready to run on the Antithesis platform!
