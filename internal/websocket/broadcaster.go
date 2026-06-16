package websocket

type Broadcaster struct {
	Hub *Hub
}

func NewBroadcaster(hub *Hub) *Broadcaster {
	return &Broadcaster{
		Hub: hub,
	}
}

func (b *Broadcaster) BroadcastSOSCreated(data interface{}) {
	b.Hub.SendToAdmins(Event{
		Type: EventSOSCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastAssistanceCreated(data interface{}) {
	b.Hub.SendToAdmins(Event{
		Type: EventAssistanceCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastReportCreated(data interface{}) {
	b.Hub.SendToAdmins(Event{
		Type: EventReportCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastAlertCreated(country string, data interface{}) {
	b.Hub.SendToCountry(country, Event{
		Type: EventAlertCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastAlertApproved(country string, data interface{}) {
	b.Hub.SendToCountry(country, Event{
		Type: EventAlertApproved,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastNotificationCreated(userID string, data interface{}) {
	b.Hub.SendToUser(userID, Event{
		Type: EventNotificationCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastChatMessage(userID string, data interface{}) {
	b.Hub.SendToUser(userID, Event{
		Type: EventChatMessageCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastSOSStatusUpdated(userID string, data interface{}) {
	b.Hub.SendToUser(userID, Event{
		Type: EventSOSStatusUpdated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastChatMessageToAdmins(data interface{}) {
	b.Hub.SendToAdmins(Event{
		Type: EventChatMessageCreated,
		Data: data,
	})
}

func (b *Broadcaster) BroadcastAssistanceStatusUpdated(
	userID string,
	data interface{},
) {
	b.Hub.SendToUser(userID, Event{
		Type: EventAssistanceStatusUpdated,
		Data: data,
	})
}