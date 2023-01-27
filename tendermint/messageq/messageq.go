package messageq

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/avast/retry-go"
)

type MessageQueue struct {
	queue          chan MessageWrapper
	processMessage func(msg []byte) (response interface{}, err error)
}

type MessageWrapper struct {
	Response chan interface{}
	Msg      []byte
}

func NewMessageQueue(processMessageFunc func(msg []byte) (response interface{}, err error)) *MessageQueue {
	return &MessageQueue{
		queue:          make(chan MessageWrapper),
		processMessage: processMessageFunc,
	}
}

func (q *MessageQueue) Add(bftTxBytes []byte) (res interface{}) {
	c := make(chan interface{})
	q.queue <- MessageWrapper{c, bftTxBytes}
	return <-c
}

func (q *MessageQueue) RunMsgEngine(num int) {
	for i := 0; i < num; i++ {
		go func() {
			for {
				m := <-q.queue
				var res interface{}
				var err error
				err = retry.Do(func() error {
					res, err = q.processMessage(m.Msg)
					if err != nil {
						log.WithError(err).Error("msgQ:processMessage")
						return err
					}
					return nil
				},
					// retry.RetryIf(func(err error) bool {
					// 	log.WithError(err).Error("msgQ:retryIf")
					// 	return true
					// }),
					retry.Attempts(6),
					retry.Delay(1*time.Second),
					retry.DelayType(retry.FixedDelay),
				)
				if err != nil {
					log.WithError(err).Errorf("could not process message in RunMsgEngine: %s", err.Error())
				}
				m.Response <- res
			}
		}()
	}
}
