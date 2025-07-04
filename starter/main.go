package main

import (
	"context"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"log"
	"temporal-sa/spiffeauth"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/helloworld"
)

func main() {
	sub, err := spiffeid.FromString("spiffe://example.org/brendan-myers.a2dd6/worker")
	if err != nil {
		log.Fatalln("unable to create SPIFFE ID: %w", err)
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:  "127.0.0.1:7233",
		Namespace: "brendan-myers.a2dd6",
		HeadersProvider: &spiffeauth.SpiffeHeadersProvider{
			Config: spiffeauth.SpiffeConfig{
				SpiffeID:   sub,
				SocketPath: "unix:///tmp/spire-agent/public/api.sock",
				Audience:   "temporal_cloud_proxy",
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "spiffe-auth",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
