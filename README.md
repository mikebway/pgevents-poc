# Demonstration EventArc Audit Log Event Handler

This simple Cloud Run container repository does nothing terribly useful, it just demonstrates that Postgres database 
record inserts and updates can give rise to [GCP EventArc](https://cloud.google.com/eventarc/docs) Pub/Sub events
and how to process those events as they are pushed to a Cloud Run service. 

For more information on how to set up the larger context for which this code is just a final element, see  
[Enable Data Access audit logs](https://cloud.google.com/logging/docs/audit/configure-data-access),
[Audit for PostgreSQL using pgAudit](https://cloud.google.com/sql/docs/postgres/pg-audit), and
[Receive a Cloud Audit Logs event](https://cloud.google.com/eventarc/docs/run/cal). 

The event receiver described in that last link is extremely crude. The [`cloud-events-sdk-eventarc`](https://github.com/salrashid123/cloud-events-sdk-eventarc)
GitHub repository demonstrates how to use the EventArc and CloudEvents libraries to do something more sophisticated.
The [`cloud-events-sdk-eventarc`](https://github.com/salrashid123/cloud-events-sdk-eventarc) code is getting a little
old and uses some deprecated CloudEvent methods; the [Golang SDK for CloudEvents](https://cloudevents.github.io/sdk-go/)
documentation straightens some of that out.

## Defining the EventArc Trigger

Here is a template for the `gcloud` command to define the trigger. Be sure to replace the `pgevents-poc`, `pg-eventing`,
and `event-source` values with the names that you have chosen for your implementation.

```shell
gcloud eventarc triggers create pgevents-poc --location=us-central1 \
  --destination-run-service=pgevents-poc \
  --destination-run-region=us-central1 \
  --event-filters="type=google.cloud.audit.log.v1.written" \
  --event-filters="resourceName=instances/event-source" \
  --event-filters="serviceName=cloudsql.googleapis.com" \
  --event-filters="methodName=cloudsql.instances.query" \
  --service-account=pg-eventing@postgres-eventing.iam.gserviceaccount.com
```

## Acknowledgements

Appreciation to [salrashid123](https://github.com/salrashid123) for their [`cloud-events-sdk-eventarc`](https://github.com/salrashid123/cloud-events-sdk-eventarc)
which clarified much around how to use the GCP EventArc Go libraries.
