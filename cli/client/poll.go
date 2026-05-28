package client

import (
	"fmt"
	"time"
)

func PollTask(client *Client, taskID string, interval time.Duration, timeout time.Duration) (*Task, error) {
	start := time.Now()

	task, err := client.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("initial fetch error: %w", err)
	}

	initialStatus := task.Status

	for {
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout exceeded (%v)", timeout)
		}

		time.Sleep(interval)

		task, err = client.GetTask(taskID)
		if err != nil {
			return nil, fmt.Errorf("poll fetch error: %w", err)
		}

		if task.Status != initialStatus {
			return task, nil
		}
	}
}