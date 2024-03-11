package blacklist

import (
	"fmt"
	"sync"
)

var blackList sync.Map

func init() {
	blackList = sync.Map{}
}

func userId2Key(id int) string {
	return fmt.Sprintf("userid_%d", id)
}

func BanUser(id int) {
	blackList.Store(userId2Key(id), true)
}

func UnbanUser(id int) {
	blackList.Delete(userId2Key(id))
}

func IsUserBanned(id int) bool {
	_, ok := blackList.Load(userId2Key(id))
	return ok
}
