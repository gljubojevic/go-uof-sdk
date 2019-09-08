package uof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageParseRoutingKeys(t *testing.T) {
	data := []struct {
		key string
		rm  Message
	}{
		{
			key: "hi.-.live.bet_cancel.21.sr:match.13073610.-",
			rm: Message{
				Type:     MessageTypeBetCancel,
				Scope:    MessageScopeLive,
				Priority: MessagePriorityHigh,
				SportID:  21,
				EventURN: "sr:match:13073610",
				EventID:  13073610,
			},
		},
		{
			key: "hi.pre.-.odds_change.1.sr:match.1234.-",
			rm: Message{
				Type:     MessageTypeOddsChange,
				Scope:    MessageScopePrematch,
				Priority: MessagePriorityHigh,
				SportID:  1,
				EventURN: "sr:match:1234",
				EventID:  1234,
			},
		},
		{
			key: "hi.virt.-.odds_change.7.vs:match.12345.-",
			rm: Message{
				Type:     MessageTypeOddsChange,
				Scope:    MessageScopeVirtuals,
				Priority: MessagePriorityHigh,
				SportID:  7,
				EventURN: "vs:match:12345",
				EventID:  12345,
			},
		},
		{
			key: "-.-.-.alive.-.-.-.-",
			rm: Message{
				Type:     MessageTypeAlive,
				Scope:    MessageScopeSystem,
				Priority: MessagePriorityLow,
			},
		},
		{
			key: "-.-.-.snapshot_complete.-.-.-",
			rm: Message{
				Type:     MessageTypeSnapshotComplete,
				Scope:    MessageScopeSystem,
				Priority: MessagePriorityLow,
			},
		},
		{
			key: "hi.-.live.odds_change.4.sr:match.11784628",
			rm: Message{
				Type:     MessageTypeOddsChange,
				Scope:    MessageScopeLive,
				Priority: MessagePriorityHigh,
				SportID:  4,
				EventURN: "sr:match:11784628",
				EventID:  11784628,
			},
		},
		{
			key: "lo.pre.live.bet_settlement.8.sr:match.12.-",
			rm: Message{
				Type:     MessageTypeBetSettlement,
				Scope:    MessageScopePrematchAndLive,
				Priority: MessagePriorityLow,
				SportID:  8,
				EventURN: "sr:match:12",
				EventID:  12,
			},
		},
	}

	for _, d := range data {
		rm, err := NewQueueMessage(d.key, nil)
		assert.Nil(t, err)
		assert.Equal(t, d.rm.Scope, rm.Scope)
		assert.Equal(t, d.rm.Type, rm.Type)
		assert.Equal(t, d.rm.Priority, rm.Priority)
		assert.Equal(t, d.rm.SportID, rm.SportID)
		assert.Equal(t, d.rm.EventURN, rm.EventURN)
		assert.Equal(t, d.rm.EventID, rm.EventID)
		assert.Equal(t, LangNone, rm.Lang)
	}
}

func TestMessageTypeParse(t *testing.T) {
	var y MessageType
	y.Parse("alive")
	assert.Equal(t, MessageTypeAlive, y)
}
