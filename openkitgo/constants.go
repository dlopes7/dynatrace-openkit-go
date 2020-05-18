package openkitgo

type EventType int

const (
	EventTypeACTION        EventType = 1
	EventTypeVALUE_STRING  EventType = 11
	EventTypeVALUE_INT     EventType = 12
	EventTypeVALUE_DOUBLE  EventType = 13
	EventTypeNAMED_EVENT   EventType = 10
	EventTypeSESSION_START EventType = 18
	EventTypeSESSION_END   EventType = 19
	EventTypeWEBREQUEST    EventType = 30
	EventTypeERROR         EventType = 40
	EventTypeCRASH         EventType = 50
	EventTypeIDENTIFY_USER EventType = 60
	EventTypeOTHER         EventType = -1
)

const (
	OPENKIT_VERSION       = "7.0.0000"
	PROTOCOL_VERSION      = 3
	PLATFORM_TYPE_OPENKIT = 1
	AGENT_TECHNOLOGY_TYPE = "ok-ext-citrix"
)
