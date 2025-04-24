## API DESIGN

- [API DESIGN](#api-design)
  - [Response Format](#response-format)
  - [VM Resource Definition](#vm-resource-definition)
  - [`PUT /vms` - Creates a new VM](#put-vms---creates-a-new-vm)
      - [Parameters (JSON body)](#parameters-json-body)
      - [Responses](#responses)
      - [Example cURL](#example-curl)
  - [`GET /vms` - List existing VMs](#get-vms---list-existing-vms)
      - [Parameters (URL Query)](#parameters-url-query)
      - [Responses](#responses-1)
      - [Example cURL](#example-curl-1)
  - [`GET /vms/{id}` - Get VM details using its defined ID](#get-vmsid---get-vm-details-using-its-defined-id)
      - [Parameters](#parameters)
      - [Responses](#responses-2)
      - [Example cURL](#example-curl-2)
  - [`DELETE /vms/{id}` - Deletes an existing VM using its defined ID](#delete-vmsid---deletes-an-existing-vm-using-its-defined-id)
      - [Parameters](#parameters-1)
      - [Responses](#responses-3)
      - [Example cURL](#example-curl-3)
  - [`POST /vms/{id}/start` - Start a VM using its defined ID](#post-vmsidstart---start-a-vm-using-its-defined-id)
      - [Parameters](#parameters-2)
      - [Responses](#responses-4)
      - [Example cURL](#example-curl-4)
  - [`POST /vms/{id}/stop` - Stop a VM using its defined ID](#post-vmsidstop---stop-a-vm-using-its-defined-id)
      - [Parameters](#parameters-3)
      - [Responses](#responses-5)
      - [Example cURL](#example-curl-5)
  - [`POST /vms/{id}/restart` - Restart a VM using its defined ID](#post-vmsidrestart---restart-a-vm-using-its-defined-id)
      - [Parameters](#parameters-4)
      - [Responses](#responses-6)
      - [Example cURL](#example-curl-6)
  - [`GET /vms/{id}/stats` - Retrieve VM usage and performance metrics using its defined ID](#get-vmsidstats---retrieve-vm-usage-and-performance-metrics-using-its-defined-id)
      - [Parameters](#parameters-5)
      - [Responses](#responses-7)
      - [Example cURL](#example-curl-7)

REST API design document for service that manages KVM virtual machines.

For all endpoints that accept a request body (e.g., `POST`), parameters are provided as JSON. All responses are returned in JSON format.

### Response Format

The general response format follow the structure below:

| Field             | Type   | Purpose                                                               |
| ----------------- | ------ | --------------------------------------------------------------------- |
| `status`          | string | Indicates the status of the response, possible values {"ok", "error"} |
| `message`         | string | A friendly message to describe the response                           |
| `item` (optional) | object | Represents information about a specific object                        |

In case of **errors**, we return a single error or list of errors depending on the endpoint, each error follow the structure below:
| Field     | Type    | Purpose                                                                                              |
| --------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `code`    | integer | This field indicates a specific error code so the client can act accordingly                         |
| `field`   | string  | (optional) In case of validation errors, this field indicates the source field that caused the error |
| `message` | string  | Reason of failure                                                                                    |

For endpoints that returned **paginated** results, `items` is returned instead of `item` and the following extra fields are available:

| Field         | Type    | Purpose                             |
| ------------- | ------- | ----------------------------------- |
| `page`        | integer | Indicates the current page number   |
| `page_count`  | integer | Indicates the number of pages       |
| `per_page`    | integer | Indicates how many item in a page   |
| `total_count` | integer | Indicates the total number of items |
| `items`       | list    | Represents a list of objects        |

### VM Resource Definition

Every VM or **domain** is defined as follow:

| Field             | Description                       | Example Value                        |
| ----------------- | --------------------------------- | ------------------------------------ |
| `id`              | VM Unique Identifier              | 56071446-7713-4cbb-ac21-9d685878b128 |
| `name`            | VM Name                           | debian-12-x64                        |
| `cpu`             | Number of CPU cores               | 2                                    |
| `memory`          | Memory allocated in MiB           | 4096                                 |
| `disk`            | Disk size in GiB                  | 60                                   |
| `cpuset`          | Core IDs for CPU pinning          | [0,4,8,12]                           |
| `read_iops_sec`   | Read IOPS                         | 300                                  |
| `write_iops_sec`  | Write IOPS                        | 150                                  |
| `total_iops_sec`  | Total IOPS for R/W                | 500                                  |
| `read_bytes_sec`  | Read bandwidth in MiB per second  | 10                                   |
| `write_bytes_sec` | Write bandwidth in MiB per second | 5                                    |
| `total_bytes_sec` | Total bandwidth in MiB for R/W    | 300                                  |

An example of a VM resource detail in an HTTP response will look like:

```json
{
    "id": "56071446-7713-4cbb-ac21-9d685878b128",
    "name": "debian-12-x64",
    "cpu": 2,
    "memory": 4096, // always in MiB
    "disk" : 60,
    "cpuset": [0,4,8,12],
    "read_iops_sec": 300,
    "write_iops_sec": 150,
    "total_iops_sec": 500,
    "read_bytes_sec": 10, // always in MiB
    "write_bytes_sec": 5 , // always in MiB
    "total_bytes_sec": 300 // always in MiB
}
```

--------------------------------------------------------------------------------

### `PUT /vms` - Creates a new VM

##### Parameters (JSON body)

> | name      |  type     | data type | description |
> |-----------|---------- |-----------|-------------|
> | cpu      | required | int ($int64) | Requested CPU cores |
> | memory   | required | int ($int64) | Requested memory size in MiB |
> | disk     | required | int ($int64) | Requested disk size in GiB |
> | read_iops_sec | required | int ($int64) | Requested read IOPS |
> | write_iops_sec| required | int ($int64) | Requested write IOPS |
> | total_iops_sec | required | int ($int64) | Requested total IOPS |
> | read_bytes_sec | required | int ($int64) | Request read bandwidth in MiB per second |
> | write_bytes_sec | required | int ($int64) | Request write bandwidth in MiB per second |
> | total_bytes_sec | required | int ($int64) | Request total bandwidth in MiB per second |

##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `202` | `application/json` | `{"status":"success","message": "VM created successfully", "data": { VMObject } }`|
> | `400` | `application/json` | `{"status":"error", "message": "VM creation failed", "errors": []`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while creating the VM", "errors": []`|

##### Example cURL

> ```javascript
>  curl -X PUT -H "Content-Type: application/json" --data @create-vm.json http://localhost:8080/vms
> ```

### `GET /vms` - List existing VMs

##### Parameters (URL Query)

> | name      |  type     | data type | description |
> |-----------|---------- |-----------|-------------|
> | state      | optional | string | Filter by state |

##### Responses

> | http code     | content-type                      | response                                                            |
> |---------------|-----------------------------------|---------------------------------------------------------------------|
> | `200`         | `application/json`                | `{"status":"success","message": "VMs enumerated successfully", "items": [{ VMObject }]}`|

##### Example cURL

> ```javascript
>  curl -X POST -H "Content-Type: application/json"  http://localhost:8080/?state=running&page=2&per_page=20
> ```

### `GET /vms/{id}` - Get VM details using its defined ID

##### Parameters

> | name |  type | data type | description |
> |------|-------|-----------|-------------|
> | None |  N/A | N/A  | N/A  |

##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `200` | `application/json` | `{"status":"success","message": "VM data returned successfully", "data": { VMObject } }`|
> | `404` | `application/json` | `{"status":"error", "message": "VM not found", "error": {}`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while getting the VM's data", "error": {}`|

##### Example cURL

> ```javascript
>  curl -X GET -H "Content-Type: application/json" http://localhost:8080/vms/56071446-7713-4cbb-ac21-9d685878b128
> ```

### `DELETE /vms/{id}` - Deletes an existing VM using its defined ID

##### Parameters

> | name |  type | data type | description |
> |------|-------|-----------|-------------|
> | None |  N/A | N/A  | N/A  |

##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `200` | `application/json` | `{"status":"success","message": "VM deleted successfully", "data": { VMObject } }`|
> | `404` | `application/json` | `{"status":"error", "message": "VM not found", "error": {}`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while deleting the VM", "error": {}`|

##### Example cURL

> ```javascript
>  curl -X DELETE -H "Content-Type: application/json" http://localhost:8080/vms/56071446-7713-4cbb-ac21-9d685878b128
> ```

### `POST /vms/{id}/start` - Start a VM using its defined ID

##### Parameters

> | name |  type | data type | description |
> |------|-------|-----------|-------------|
> | None |  N/A | N/A  | N/A  |


##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `200` | `application/json` | `{"status":"success","message": "VM started successfully", "data": { VMObject } }`|
> | `404` | `application/json` | `{"status":"error", "message": "VM not found", "error": {}`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while starting the VM", "error": {}`|

##### Example cURL

> ```javascript
>  curl -X POST -H "Content-Type: application/json" http://localhost:8080/vms/56071446-7713-4cbb-ac21-9d685878b128/start
> ```

### `POST /vms/{id}/stop` - Stop a VM using its defined ID

##### Parameters

> | name |  type | data type | description |
> |------|-------|-----------|-------------|
> | None |  N/A | N/A  | N/A  |

##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `200` | `application/json` | `{"status":"success","message": "VM stopped successfully", "data": { VMObject } }`|
> | `404` | `application/json` | `{"status":"error", "message": "VM not found", "error": {}`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while stopping the VM", "error": {}`|

##### Example cURL

> ```javascript
>  curl -X POST -H "Content-Type: application/json" http://localhost:8080/vms/56071446-7713-4cbb-ac21-9d685878b128/stop
> ```

### `POST /vms/{id}/restart` - Restart a VM using its defined ID

##### Parameters

> | name |  type | data type | description |
> |------|-------|-----------|-------------|
> | None |  N/A | N/A  | N/A  |

##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `200` | `application/json` | `{"status":"success","message": "VM restarted successfully", "data": { VMObject } }`|
> | `404` | `application/json` | `{"status":"error", "message": "VM not found", "error": {}`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while restarting the VM", "error": {}`|

##### Example cURL

> ```javascript
>  curl -X POST -H "Content-Type: application/json" http://localhost:8080/vms/56071446-7713-4cbb-ac21-9d685878b128/restart
> ```

### `GET /vms/{id}/stats` - Retrieve VM usage and performance metrics using its defined ID

##### Parameters

> | name |  type | data type | description |
> |------|-------|-----------|-------------|
> | None |  N/A | N/A  | N/A  |

##### Responses

> | http code | content-type | response |
> |-----------|--------------|----------|
> | `200` | `application/json` | `{"status":"success","message": "VM restarted successfully", "data": { VMObject } }`|
> | `404` | `application/json` | `{"status":"error", "message": "VM not found", "error": {}`|
> | `500` | `application/json` | `{"status":"error", "message": "Something went wrong while retrieving the VM's statistics", "error": {}`|

##### Example cURL

> ```javascript
>  curl -X GET -H "Content-Type: application/json" http://localhost:8080/vms/56071446-7713-4cbb-ac21-9d685878b128/stats
> ```
