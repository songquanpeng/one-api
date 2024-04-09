package notify

var notifyChannels = New()

type Notify struct {
	notifiers map[string]Notifier
}

func (n *Notify) addChannel(channel Notifier) {
	if channel != nil {
		channelName := channel.Name()
		if _, ok := n.notifiers[channelName]; ok {
			return
		}
		n.notifiers[channelName] = channel
	}
}

func (n *Notify) addChannels(channel ...Notifier) {
	for _, s := range channel {
		n.addChannel(s)
	}
}

func New() *Notify {
	notify := &Notify{
		notifiers: make(map[string]Notifier, 0),
	}

	return notify
}

func AddNotifiers(channel ...Notifier) {
	notifyChannels.addChannels(channel...)
}
