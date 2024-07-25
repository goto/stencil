# Detect Schema Change

## Enabling Schema Change

To enable schema change toggle on the flag `SCHEMACHANGE_ENABLE` and provide `KAFKAPRODUCER_BOOTSTRAPSERVER`
and `SCHEMACHANGE_KAFKATOPIC`.

## Create a schema

```bash
# create namespace named "quickstart"
curl -X POST http://localhost:8000/v1beta1/namespaces -H 'Content-Type: application/json' -d '{"id": "quickstart", "format": "FORMAT_PROTOBUF", "compatibility": "COMPATIBILITY_BACKWARD", "description": "This field can be used to store namespace description"}'

# create descriptor file v1
protoc --descriptor_set_out=./protos/1.desc --include_imports ./guide/protos/1.proto

It will be version 1
# upload generated proto descriptor file to server with schema name as `example` under `quickstart` namespace.
curl -H "X-SourceURL:www.github.com/some-repo" -H "X-CommitSHA:some-commit-sha" -X POST http://localhost:8000/v1beta1/namespaces/quickstart/schemas/example --data-binary "@1.desc"

# create descriptor file v2
protoc --descriptor_set_out=./protos/1.desc --include_imports ./guide/protos/2.proto

It will be version 2

# upload generated proto descriptor file to server with schema name as `example` under `quickstart` namespace.
curl -X POST http://localhost:8000/v1beta1/namespaces/quickstart/schemas/example --data-binary "@2.desc"

# Detect Schema Change
Now, when the schema is uploaded in DB. Below things happened parallely in separate go routine
1. It will get the previous schema data corresponding to that namespace and schema name.
2. It will check all the messages in that schema whether they are impacted(directly/indireclty) or not. 
3. For change, it will check if any fileds got added/deleted. In this case only `address` field is added in `friend` message. So, `friend` proto is only impacted.
4. In this way it gets all impacted messages. And do below steps for all messages.
5. Then it will get the all indirectly impacted messages due to the message `friend`. 
6. So, it will get `User` as it is using `Friend` message as compostion.
7. It will do the same thing for `User` also until all impacted messages got proccessed.
8. Then it will make `SchemChangedEvent` object and push it to a kafka topic.
9. After that it will update the `notification_events` database table with `success=true` after pushing the object to kafka. 
```

## Handling Failures

In case of any failure while publishing to Kafka or other issues, the `notification_events` table can be utilized to
manage these failures. If the `success` field is `false` for a specific `namespace_id`, `schema_id`, and `version_id`,
it indicates that the notification was not sent. In such scenarios, the `reconciliation API` can be used to resend the
notifications.

## Depth

In schema changes, we use the `depth` parameter to determine the level of impacted schemas to retrieve. If `depth` is
set to an empty string `""` or `-1`, it will fetch all indirectly impacted schemas. However, if a specific depth is
provided, it will only retrieve schemas up to that specified level. For example, if `depth` is set to `3`, it will fetch
up to three levels of impacted schemas, even if there are more levels of actual impacted schemas.
E.g. If depth is `2` then the `impacted_schemas` for `Address` in below sample would contain only `Address` and `Friend` message.

## Sample SchemaChangedEvent Object

```
{
  "namespace_name": "quickstart",
  "schema_name": "example",
  "updated_schemas": [
    "test.stencil.main.Friend",
    "test.stencil.main.Address",
  ],
  "updated_fields": {
    "test.stencil.main.Friend": {
      "field_names": [
        "address"
      ]
    },
    "test.stencil.main.Address": {
      "field_names": [
        "city"
      ]
    }
  },
  "impacted_schemas": {
    "test.stencil.main.Address": {
      "schema_names": [
        "test.stencil.main.Address",
        "test.stencil.main.Friend",
        "test.stencil.main.User",
      ]
    },
    "test.stencil.main.Friend": {
      "schema_names": [
        "test.stencil.main.Friend",
        "test.stencil.main.User",
      ]
    }
  },
  "version": 1,
  "metadata": {
    "source_url": "https://github.com/some-repo",
    "commit_sha": "some-commit-sha"
  }
}
```

## Showing MR information 
While calculating the impacted schemas, the `SchemaChangedEvent` will also include information about the source URL and commit SHA. The `source_url` represents the repository URL, and the `commit_sha` corresponds to the commit SHA associated with that version.

## Detect Schema Change
```bash
$ curl -X GET http://localhost:8080/v1beta1/schema/detect-change/quickstart/example?from=1&to=2;
```