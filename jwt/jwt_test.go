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

	jwt2 := BuildJWTWithTime("user id 2", "", time.Second)
	time.Sleep(2 * time.Second)
	ctx2, ok = check(ctx, jwt2)
	assert.NotNil(t, ctx2)
	assert.False(t, ok)
}

func TestValidateUser(t *testing.T) {
	user := "user-name"
	wrong := "wrong-name"
	jwt := BuildJWT(user)
	assert.Equal(t, user, validateAndUser(jwt))
	assert.NotEmpty(t, validateAndUser(jwt))
	a := []string{"Bearer " + jwt}
	ok := IsValidUser(user, a)
	assert.True(t, ok)
	ok = IsValidUser(wrong, a)
	assert.False(t, ok)
}

func TestCalcExpiration(t *testing.T) {
	jwt := BuildJWT("user-name")
	yes := IsTimeLeft(jwt, 24*time.Hour)
	assert.True(t, yes)
	yes = IsTimeLeft(jwt, 3*24*time.Hour+time.Second)
	assert.False(t, yes)
}
