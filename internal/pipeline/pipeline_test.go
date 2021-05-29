package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRestartTask(t *testing.T) {
	assert := assert.New(t)

	id := "my-id"
	initialUrl := "https://example.com/initial"
	restartUrl := "https://example.com/restart"

	count0 := 0
	count1 := 0

	f0 := func(ctx context.Context, tk *Task) error {
		count0++
		if tk.URL == initialUrl {
			return tk.Restart(ctx, restartUrl)
		}
		return nil
	}

	f1 := func(ctx context.Context, tk *Task) error {
		count1++
		return nil
	}

	p := BuildPipeline(f0, f1)

	task := NewTask(nil, id, initialUrl, p)
	err := task.Run(context.TODO())
	assert.Nil(err)

	// The first pipeline step should have run in both iterations,
	// the second step in the restarted part only
	assert.Equal(2, count0)
	assert.Equal(1, count1)

	assert.Equal(restartUrl, task.URL)
}
