package main

import (
    "strings"
)

/*
 * List of users in an IRC channel.
 */
type UserManager struct {
    users map[string] bool
}

func NewUserManager() *UserManager {
    users := make(map[string] bool)
    return &UserManager{users}
}

// User leaves a channel
func (self *UserManager) Remove(user string) bool {
    if !self.users[user] {
        return false
    }
	self.users[user] = false, false
    return true
}

// User joins a channel
func (self *UserManager) Add(user string) {
	if strings.HasPrefix(user, "@") || strings.HasPrefix(user, "+") {
		user = user[1:]
	}
	self.users[user] = true
}

// Replaces current list of users with that given
func (self *UserManager) Update(users []string) {
	self.users = make(map[string] bool)
	for _, user := range users {
		self.Add(user)
	}
}

// Is this user in our channel?
func (self *UserManager) Has(user string) bool {
    result := self.users[user]
    rawLog.Println("Has: " + user, result)
    return result
}

// First nick which starts with 'prefix', or 'prefix' if no match
func (self *UserManager) FirstMatch(prefix string) string {
    if len(prefix) == 0 {
        return prefix
    }

    for key, _ := range self.users {
        if strings.HasPrefix(key, prefix) {
            return key
        }
    }
    return prefix
}
