package main

import (
	"time"

	"github.com/gentlemanautomaton/filehealth"
	"github.com/gentlemanautomaton/volmgmt/fileattr"
)

func buildHandlers() []filehealth.IssueHandler {
	now := time.Now()
	return []filehealth.IssueHandler{
		filehealth.AttrHandler{Unwanted: fileattr.Temporary},
		filehealth.TimeHandler{Max: now, Reference: now, Lenience: time.Hour * 24},
		filehealth.NameHandler{TrimSpace: true},
	}
}
