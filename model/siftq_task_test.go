//go:build cgo

package model

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSiftQTaskBindingsAreUserScoped(t *testing.T) {
	oldDB := DB
	t.Cleanup(func() { DB = oldDB })
	var err error
	DB, err = gorm.Open(sqlite.Open("file:siftq-bindings?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := DB.AutoMigrate(&SiftQTaskBinding{}); err != nil {
		t.Fatal(err)
	}

	bindings := []*SiftQTaskBinding{
		{TaskID: "task-a", ChannelID: 52, UserID: 1, TokenID: 11},
		{TaskID: "task-a", ChannelID: 53, UserID: 2, TokenID: 22},
		{TaskID: "old-task", ChannelID: 52, UserID: 1, TokenID: 11, CreatedAt: time.Now().Add(-8 * 24 * time.Hour).Unix()},
	}
	for _, binding := range bindings {
		if err := SaveSiftQTaskBinding(binding); err != nil {
			t.Fatal(err)
		}
	}

	userOne, err := GetSiftQTaskBinding(1, "task-a")
	if err != nil || userOne.ChannelID != 52 {
		t.Fatalf("user one binding: %#v, %v", userOne, err)
	}
	userTwo, err := GetSiftQTaskBinding(2, "task-a")
	if err != nil || userTwo.ChannelID != 53 {
		t.Fatalf("user two binding: %#v, %v", userTwo, err)
	}

	listed, err := ListSiftQTaskBindings(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].TaskID != "task-a" {
		t.Fatalf("unexpected list: %#v", listed)
	}

	if err := DeleteSiftQTaskBinding(1, "task-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := GetSiftQTaskBinding(1, "task-a"); err == nil {
		t.Fatal("deleted binding still exists")
	}
	if _, err := GetSiftQTaskBinding(2, "task-a"); err != nil {
		t.Fatal("other user's binding was deleted")
	}
}
