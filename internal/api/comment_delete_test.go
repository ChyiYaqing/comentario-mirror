package api

import (
	"testing"
	"time"
)

func TestCommentDeleteBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	commenterHex := "temp-commenter-hex"
	commentHex, _ := commentNew(commenterHex, "example.com", "/path.html", "root", "**foo**", "approved", time.Now().UTC())
	_, _ = commentNew(commenterHex, "example.com", "/path.html", commentHex, "**bar**", "approved", time.Now().UTC())

	if err := commentDelete(commentHex, commenterHex); err != nil {
		t.Errorf("unexpected error deleting comment: %v", err)
		return
	}

	c, _, _ := commentList(commenterHex, "example.com", "/path.html", false)

	if len(c) != 0 {
		t.Errorf("expected no comments found %d comments", len(c))
		return
	}
}

func TestCommentDeleteEmpty(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if err := commentDelete("", "test-commenter-hex"); err == nil {
		t.Errorf("expected error deleting comment with empty commentHex")
		return
	}
}
