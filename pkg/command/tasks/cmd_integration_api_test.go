// Copyright (C) 2024 ScyllaDB

//go:build all || api_integration
// +build all api_integration

package tasks

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"testing"

	"github.com/scylladb/scylla-manager/v3/pkg/managerclient"
	"github.com/scylladb/scylla-manager/v3/swagger/gen/scylla-manager/models"
)

const (
	authToken        = "token"
	clusterIntroHost = "192.168.200.11"
)

func TestSctoolTasksLabelsIntegrationAPITest(t *testing.T) {
	client, err := managerclient.NewClient("http://localhost:5080/api/v1")
	if err != nil {
		t.Fatalf("Unable to create managerclient to consume manager HTTP API, err = {%v}", err)
	}

	clusterID, err := client.CreateCluster(context.Background(), &models.Cluster{
		AuthToken: authToken,
		Host:      clusterIntroHost,
	})
	if err != nil {
		t.Fatalf("Unable to create cluster for further listing, err = {%v}", err)
	}

	defer func() {
		if err := client.DeleteCluster(context.Background(), clusterID); err != nil {
			t.Fatalf("Failed to delete cluster, err = {%v}", err)
		}
	}()

	taskID, err := client.CreateTask(context.Background(), clusterID, &managerclient.Task{
		Type:    "repair",
		Enabled: true,
		Labels: map[string]string{
			"k1": "v1",
		},
		Properties: make(map[string]interface{}),
	})
	if err != nil {
		t.Fatalf("Failed to create task, err = {%v}", err)
	}

	if err := client.UpdateTask(context.Background(), clusterID, &managerclient.Task{
		ID:      taskID.String(),
		Type:    "repair",
		Enabled: true,
		Labels: map[string]string{
			"k2": "v2",
		},
		Properties: make(map[string]interface{}),
	}); err != nil {
		t.Fatalf("Unable to update cluster for further listing, err = {%v}", err)
	}

	cmd := exec.Command("./sctool.api-tests", "tasks", "--cluster", clusterID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Dir = "/scylla-manager"

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unable to list tasks with sctool tasks, err = {%v}, stderr = {%v}", err, stderr.String())
	}

	re := regexp.MustCompile(fmt.Sprintf(`%s *\| *k2=v2`, taskID))
	if !re.Match(output) {
		t.Fatalf("Expected to get pattern {<task_id> *\\| *k2=v2}, got {%s}", string(output))
	}
}
