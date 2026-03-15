package audit

import "log"

//	{
//		"ts": 12345678,        // unix timestamp события
//		"action": "shorten",   // действие: shorten (создание) или follow (прохождение по ссылке)
//		"user_id": "12315134", // идентификатор пользователя, если есть
//		"url": "https://mylongdomain.com/my/long/path/to/shorten/" // оригинальный (не сокращенный) URL
//	}
type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id"`
	URL       string `json:"url"`
}

type Service struct {
	observers []Observer
}

func NewService() *Service {
	return &Service{
		observers: make([]Observer, 0),
	}
}

func (a *Service) AddObserver(obs Observer) {
	a.observers = append(a.observers, obs)
}

func (a *Service) NotifyAll(event Event) {
	for _, observer := range a.observers {
		obs := observer
		go func() {
			if err := obs.Notify(event); err != nil {
				log.Printf("audit notify failed: %v", err)
			}

		}()
	}
}
