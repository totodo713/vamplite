package ecs

import (
	"errors"
)

// EventBusImpl - 未実装のEventBus実装（TDD Red段階用）
type EventBusImpl struct {
	// すべてのフィールドは後で実装予定
}

// NewEventBus creates a new EventBus instance (未実装)
func NewEventBus(config *EventBusConfig) EventBus {
	return &EventBusImpl{}
}

// 全メソッドを未実装状態で定義（TDD Red段階）

func (eb *EventBusImpl) Start() error {
	return errors.New("not implemented")
}

func (eb *EventBusImpl) Stop() error {
	return errors.New("not implemented")
}

func (eb *EventBusImpl) IsRunning() bool {
	return false
}

func (eb *EventBusImpl) Flush() error {
	return errors.New("not implemented")
}

func (eb *EventBusImpl) Publish(eventType EventTypeID, event EventBusEvent) error {
	return errors.New("not implemented")
}

func (eb *EventBusImpl) PublishAsync(eventType EventTypeID, event EventBusEvent) error {
	return errors.New("not implemented")
}

func (eb *EventBusImpl) Subscribe(eventType EventTypeID, handler EventBusHandler) (SubscriptionID, error) {
	return 0, errors.New("not implemented")
}

func (eb *EventBusImpl) Unsubscribe(subscriptionID SubscriptionID) error {
	return errors.New("not implemented")
}

func (eb *EventBusImpl) SubscribeWithFilter(eventType EventTypeID, filter EventBusFilter, handler EventBusHandler) (SubscriptionID, error) {
	return 0, errors.New("not implemented")
}

func (eb *EventBusImpl) GetStats() EventBusStats {
	return EventBusStats{}
}

func (eb *EventBusImpl) GetSubscriptions() map[SubscriptionID]*EventBusSubscription {
	return make(map[SubscriptionID]*EventBusSubscription)
}
