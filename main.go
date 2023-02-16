package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// TODO: Unit test this code!!!

// EventSummaryLogPayload is used to render many of a CloudEvent's attributes for structured logging
type EventSummaryLogPayload struct {

	// SpecVersion returns the native CloudEvents Spec version of the event context.
	SpecVersion string

	// Type returns the CloudEvents type from the context.
	Type string

	// Source returns the CloudEvents source from the context.
	Source string

	// Subject returns the CloudEvents subject from the context.
	Subject string

	// ID returns the CloudEvents ID from the context.
	ID string

	// Time returns the CloudEvents creation time from the context.
	Time time.Time

	// DataSchema returns the CloudEvents schema URL (if any) from the
	// context.
	DataSchema string

	// DataContentType returns content type on the context.
	DataContentType string

	// DeprecatedDataContentEncoding returns content encoding on the context.
	DeprecatedDataContentEncoding string

	// DataMediaType returns the MIME media type for encoded data, which is
	// needed by both encoding and decoding. This is a processed form of
	// DataContentType and it may return an error.
	DataMediaType string
}

const (
	// logName is the name that we set for our GCP logger
	logName = "pgevents-crude-poc"

	// auditLogEventType is the only type of EventArc event that we handle. We check that he events that we
	// receive match this type
	auditLogEventType = "google.cloud.audit.log.v1.written"
)

var (
	// projectID is, you will never guess, the ID of the project this code gets deployed to
	projectID string

	// logClient is our GCP structured logging API client
	logClient *logging.Client

	// logger is the structured Go library logging interface that we obtain from the logClient.
	logger *logging.Logger

	// stdLog is the standard Go library logging interface that we obtain from the logClient.
	stdLog *log.Logger
)

// init is the static initializer used to configure our local and global static variables.
func init() {

	// Get the ID of the project that we have been deployed to from the Cloud Run metadata server
	var err error
	projectID, err = metadata.ProjectID()
	if err != nil {
		log.Fatalf("failed to obtain project ID from metadata: %v", err)
	}

	// Initialize our GCP logging client
	ctx := context.Background()
	logClient, err = logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("failed to obtain logging client: %v", err)
	}

	// Obtain both a google structured logging API interface and a standard Go library interface from the same logging client
	logger = logClient.Logger(logName)
	stdLog = logger.StandardLogger(logging.Info)
}

// main is the service entry point. It establishes the HTTP server and routes root path requests
// the PostgresEvents function.
func main() {

	// Obtain a CloudEvents client
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		stdLog.Printf("failed to create client, %v", err)
	}

	// Start the HTTP server to route events to our receiver
	fmt.Println("Starting Server")
	err = c.StartReceiver(context.Background(), AuditEventReceiver)
	if err != nil {
		stdLog.Printf("failed to start event receiver, %v", err)
	}
}

// AuditEventReceiver is invoked for every EventArc event that is pushed to this Cloud Run instance.
func AuditEventReceiver(ctx context.Context, event cloudevents.Event) error {

	// Log a summary of the event. In a production environment, we would not care about all of this but
	// the purpose of this code is discovery so we will log everything we can
	eventSummary := EventSummaryLogPayload{
		SpecVersion:                   event.SpecVersion(),
		Type:                          event.Type(),
		Source:                        event.Source(),
		Subject:                       event.Subject(),
		ID:                            event.ID(),
		Time:                          event.Time(),
		DataSchema:                    event.DataSchema(),
		DataContentType:               event.DataContentType(),
		DeprecatedDataContentEncoding: event.DeprecatedDataContentEncoding(),
		DataMediaType:                 event.DataMediaType(),
	}
	logger.Log(logging.Entry{Payload: eventSummary})

	// TODO: Confirm whether returning an error causes a retry - we don't want retries of errors that cannot fix themselves

	// Confirm that we have been given an event of the type we know how to handle
	if event.Type() == auditLogEventType {

		// Ask for the protobuf data of the event in GCP Cloud Logging format
		auditData := &logging.Entry{}
		if err := event.DataAs(auditData); err != nil {

			// Report that we failed to unpack the audit event data
			err = fmt.Errorf("failed to render audit data: %w", err)
			logger.Log(logging.Entry{Payload: err})
			return err
		}

		// For the purposes of the POC, just loging the audit event is sufficient
		logger.Log(logging.Entry{Payload: auditData})

	} else {

		// Report that we receive and invalid event type
		err := fmt.Errorf("invalid event type: %s", event.Type())
		logger.Log(logging.Entry{Payload: err})
		return err
	}

	// All is well!
	return nil
}
