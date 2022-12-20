# hpdrive
The file service stores objects to local storage.

## Getting started
Set up a local environemnt with Docker.
```console
# Generate TLS certificate and private key via OpenSSL before launch the server connection.
$ make certs SERVICE_NAME=localhost
# Run the `docker-compose` to start web server.
$ docker-compose up -d
```

## Implementation
Design a simple workflow that applies golang web server and sqlite database to meet the requirement, implement a [layered architectutre](https://martinfowler.com/bliki/PresentationDomainDataLayering.html) to isolate the domain model of business logic and eliminate the dependency on infrastructure, user interface or application logic (since the requirement isn't complex, so we don't consider the application layer at this stage). Define the user interface layer (controller) that is responsible for presenting information to the user and interpreting user commands, and design an infrastructure layer that access external service like database for persistence usage, the reason of database chosen is to [perform faster read/write I/O](https://www.sqlite.org/fasterthanfs.html).

Table schema: `files`
| Column       | Type         | Description                             |
| ------------ | ------------ | --------------------------------------- |
| id           | Integer      | The unique key of file ID.              |
| dir          | Varchar(100) | The directory contains a list of files. |
| fileName     | Varchar(100) | The file name.                          |
| size         | Unsigned Mediuminteger | The file size in bytes.       |
| content      | BLOB         | The file binary data.                   |
| isArchived   | TINYINT      | Determine whether the file is archived. |
| createdAt    | DATETIME     | The created date and time.              |
| lastModified | DATETIME     | The updated date and time.              |

## API reference docs
List API endpoints for file storage manipulation.

### GET `/file/{filePath}`
Retrieve the specific file in binary content or a list of files and sub-directories.

#### Path Parameters
| Field      | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| filePath   | String | Y        | A field represents the directory or file path for storing objects, list the examples as follows:<ul><li>/file/root/usr/aabb (directory)</li><li>/file/root/usr/aabb/test.txt (file)</li></ul> |

#### Query Parameters
| Field          | Type   | Required | Description |
| -------------- | ------ | -------- | ----------- |
| orderBy        | String |          | The column sorts a list of file names in order.<p>Note: both `orderBy` and `orderDirection` must be required for order sorting otherwise received an error.</p> Allow the following columns for sorting. <ul><li>`fileName`</li><li>`size`</li><li>`lastModified`</li></ul> |
| orderDirection | String |          | The field sort data in ascending or descending order.<p>Note: both `orderBy` and `orderDirection` must be required for order sorting otherwise received an error.</p> Allow following columns <ul><li>`Ascending`</li><li>`Descending`</li></ul> |
| filterByName   | String |          | The field to filter the rows by the file name. |

#### Responses
`200`
- If `filePath` is a file path, respond with binary message.
- If `{filePath}` is a directory

| Field       | Type   | Required | Description |
| ----------- | ------ | -------- | ----------- |
| isDirectory | String | Y        | Determine whether the file is a directory. |
| files       | String | Y        | A list of files contain file names and sub-directories. |

`400`, `404`, `500`
| Field   | Type   | Required | Description        |
| ------- | ------ | -------- | ------------------ |
| code    | Int    | Y        | The error code.    |
| message | String | Y        | The error message. |

```sh
GET /file/foo

HTTP/1.1 200 OK
{
  "isDirectory": true,
  "files": [
    "build.sh",
    ".gitignore"
  ]
}

GET /file/foo?orderBy=filterName

HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "Both [orderBy] and [orderDirection] fields must be specified"
}
```

### POST `/file/{filePath}`
Create a new file and store the binary content to local storage.

#### Path Parameters
| Field      | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| filePath   | String | Y        | A field represents the directory or file path for storing objects, list the examples as follows:<ul><li>/file/root/usr/aabb (directory)</li><li>/file/root/usr/aabb/test.txt (file)</li></ul> |

#### Body Parameters
| Field | Type   | Required | Description |
| ----- | ------ | -------- | ----------- |
| file  | String | Y        | The form data is stored in binary file content. |

#### Responses
`200` (Empty response)

`400`, `404`, `500`
| Field   | Type   | Required | Description        |
| ------- | ------ | -------- | ------------------ |
| code    | Int    | Y        | The error code.    |
| message | String | Y        | The error message. |

```sh
POST /file/foo/bar.txt

HTTP/1.1 200 OK
{
}

POST /file/foo

HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "file /file/foo must be a file"
}
```

### PATCH `/file/{filePath}`
Update an already existing file.

#### Path Parameters
| Field      | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| filePath   | String | Y        | A field represents the directory or file path for storing objects, list the examples as follows:<ul><li>/file/root/usr/aabb (directory)</li><li>/file/root/usr/aabb/test.txt (file)</li></ul> |

#### Body Parameters
| Field | Type   | Required | Description |
| ----- | ------ | -------- | ----------- |
| file  | String | Y        | The form data is stored in binary file content. |

#### Responses
`200` (Empty response)

`400`, `404`, `500`
| Field   | Type   | Required | Description        |
| ------- | ------ | -------- | ------------------ |
| code    | Int    | Y        | The error code.    |
| message | String | Y        | The error message. |

```sh
PATCH /file/foo/bar.txt

HTTP/1.1 200 OK
{
}

PATCH /file/foo.txt

HTTP/1.1 404 Not Found
{
  "code": 404,
  "message": "file /foo.txt does not exist"
}
```

### DELETE `/file/{filePath}`
Update an already existing file.

#### Path Parameters
| Field      | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| filePath   | String | Y        | A field represents the directory or file path for storing objects, list the examples as follows:<ul><li>/file/root/usr/aabb (directory)</li><li>/file/root/usr/aabb/test.txt (file)</li></ul> |

#### Responses
`200` (Empty response)

`400`, `500`
| Field   | Type   | Required | Description        |
| ------- | ------ | -------- | ------------------ |
| code    | Int    | Y        | The error code.    |
| message | String | Y        | The error message. |

```sh
DELETE /file/foo/bar.txt

HTTP/1.1 200 OK
{
}

DELETE /file/foo

HTTP/1.1 400 Bad Request
{
  "code": 400,
  "message": "file /file/foo must be a file"
}
```

### Epilogue
There're few things that need some improvements.
- URI naming convension: The requirement doesn't fully follow REST API standard. Generally, URI naming should start with the name (e.g. **api**), and then followed by a version number. Most importantly, URI stands for resource collections that are globally unique and hierarchical (e.g. `/files` represents a file resource, `/files/{id}/directories` means to retrieve sub-resources directories that are dependent on parent resource `files`), in this case, it's recommended to change `localSystemFilePath` to query parameter.
- HTTP verbs: Since `PATCH` is not required to be `idempotent` that used to applied partial modifications without changing a resource might result with different resource state, replace `PATCH` with `PUT` verb instead.
- Field confusion: It seems that we've put directory and file together in the path `localSystemFilePath` that might cause confusion and increase complexity since they should contain different attributes and behaviors (same as HTTP GET response), separate them into `dir` and `fileName` instead.
- Redundant fields: There're few fields seems redudant to the endpoint, POST/PATCH/DELETE `localSystemFilePath` (since it can be parsed by the form data), `isDirectory` (only response when the `localSystemFilePath` is a directory).
- Simple scenarios: The given use cases/constraints seem too simple to the assumptions so it cannot be designed for a high level architecture (choose SQL/NoSQL, split web server into read/write for individual throughput, linked-list data structure for directories, scale app strategies).
