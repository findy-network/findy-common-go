package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/lainio/err2/assert"
)

func TestBuildJWT(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	jwt := BuildJWT("user id 1")
	ctx := context.Background()
	assert.INotNil(ctx)

	ctx2, ok := check(ctx, jwt)
	assert.INotNil(ctx2)
	assert.That(ok)

	jwt2 := BuildJWTWithTime("user id 2", "", time.Second)
	time.Sleep(2 * time.Second)
	ctx2, ok = check(ctx, jwt2)
	assert.INotNil(ctx2)
	assert.ThatNot(ok)
}

func TestValidateUser(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	user := "user-name"
	wrong := "wrong-name"
	jwt := BuildJWT(user)
	assert.Equal(user, validateAndUser(jwt))
	assert.NotEmpty(validateAndUser(jwt))
	a := []string{"Bearer " + jwt}
	ok := IsValidUser(user, a)
	assert.That(ok)
	ok = IsValidUser(wrong, a)
	assert.ThatNot(ok)
}

func TestCalcExpiration(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	jwt := BuildJWT("user-name")
	yes := IsTimeLeft(jwt, 24*time.Hour)
	assert.That(yes)
	yes = IsTimeLeft(jwt, 3*24*time.Hour+time.Second)
	assert.ThatNot(yes)
}
