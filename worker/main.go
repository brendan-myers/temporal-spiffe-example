package main

import (
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"log"
	"temporal-sa/spiffeauth"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/helloworld"
)

func main() {
	sub, err := spiffeid.FromString("spiffe://example.org/brendan-myers.a2dd6/worker")
	if err != nil {
		log.Fatalln("unable to create SPIFFE ID: %w", err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
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

	w := worker.New(c, "spiffe-auth", worker.Options{})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
