package main

//go:generate msgp

import (
	//"github.com/tinylib/msgp/msgp"
)


type Msg struct {
	Ts			int64 `msg:"Ts"`
	MsgID		string `msg:"MsgID"`
	Data		string `msg:"Data"`
}