# bukky

This is an in-memory key-value store that de-duplicates data on the bucket level. Its main purpose is to show how to write a service in Go.

## Usage

`bukky` provides an HTTP server with the following endpoints:

|                           Path |   Method | Description                                                                                                                 |
|-------------------------------:|---------:|:----------------------------------------------------------------------------------------------------------------------------|
|                      `/health` |      any | Health-check which always returns `HTTP 200`. For testing if the service is running.                                        |
|                       `/stats` |    `GET` | Returns statistics about the number of buckets and objects in memory.                                                       |
| `/objects/{bucket}/{objectID}` |    `GET` | Returns the object with the specified ID saved to that bucket. If the object does not exist an `HTTP 404` is returned.      |
| `/objects/{bucket}/{objectID}` |    `PUT` | Saves the data in the request body as the specified object in that bucket. Returns `HTTP 201` and the object ID on success. |
| `/objects/{bucket}/{objectID}` | `DELETE` | Deletes the specified object from the bucket. Returns `HTTP 204` on success or `HTTP 404` if the object was not found.      |

The service is configured using these environment variables:

|          Name | Description                                                                       |
|--------------:|:----------------------------------------------------------------------------------|
| `LISTEN_ADDR` | Sets the address and port the service should be listening on. Defaults to `:8080` |
