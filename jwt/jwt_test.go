package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildJWT(t *testing.T) {
	jwt := BuildJWT("user id 1")
	ctx := context.Background()
	assert.NotNil(t, ctx)

	ctx2, ok := check(ctx, jwt)
	assert.NotNil(t, ctx2)
	assert.True(t, ok)

	timeValid = time.Second

	jwt2 := BuildJWT("user id 2")
	time.Sleep(2 * time.Second)
	ctx2, ok = check(ctx, jwt2)
	assert.NotNil(t, ctx2)
	assert.False(t, ok)
}
