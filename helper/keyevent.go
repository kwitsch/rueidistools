package helper

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/rueidis"
)

const (
	nkeKey = "notify-keyspace-events"
)

// Enables expired notify-keyspace-events if not enabled
func EnableExpiredNKE(ctx context.Context, rc rueidis.Client) error {
	needUpdate := false

	nke, err := GetNKE(ctx, rc)
	if err != nil {
		return err
	}

	if !strings.Contains(nke, "K") {
		nke = fmt.Sprintf("K%s", nke)

		needUpdate = true
	}

	if !strings.Contains(nke, "x") && !strings.Contains(nke, "A") {
		nke = fmt.Sprintf("%sx", nke)

		needUpdate = true
	}

	if needUpdate {
		err = rc.Do(ctx, rc.B().
			ConfigSet().
			ParameterValue().
			ParameterValue(nkeKey, nke).
			Build()).Error()

		if err != nil {
			return err
		}
	}

	return nil
}

// Read notify-keyspace-events config
func GetNKE(ctx context.Context, rc rueidis.Client) (string, error) {
	res, err := rc.Do(ctx, rc.B().ConfigGet().Parameter(nkeKey).Build()).AsStrMap()
	if err != nil {
		return "", err
	}

	return res[nkeKey], nil
}
