package openai

import (
	"crypto/rand"
	"fmt"
)

func genID(prefix string) string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s%x%x%x%x%x", prefix, b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func GenChatID() string     { return genID("chatcmpl-") }
func GenRespID() string     { return genID("resp_") }
func GenMsgID() string      { return genID("msg_") }
func GenFuncCallID() string { return genID("fc_") }
func GenReasoningID() string { return genID("rs_") }