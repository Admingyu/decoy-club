package users

import (
	"testing"
	"time"
)

func TestUserCountersDefaultToZero(t *testing.T) {
	u := User{
		Username:  "bot_alice",
		CreatedAt: time.Now(),
	}

	if u.PostCount != 0 || u.ReplyCount != 0 || u.FollowersCount != 0 {
		t.Fatalf("expected zero-value counters, got %+v", u)
	}
}
