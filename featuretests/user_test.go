package featuretests

import (
	"fmt"
	"testing"
	"time"

	"github.com/canthefason/irc-k/client"
	"github.com/canthefason/irc-k/common"
	"github.com/canthefason/irc-k/config"
	"github.com/canthefason/irc-k/feeder"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMessageHandling(t *testing.T) {
	go feeder.Run(config.Conf.IRC.Server)

	beaker := client.NewSubscriber()
	chef := new(client.Connection)
	chef.Server = config.Conf.IRC.Server

	Convey("While system is up and running", t, func() {
		Convey("Beaker should be able to receive messages from subscribed channel", func() {
			Convey("Beaker should be able to join channel", func() {
				err := beaker.Subscribe("muppetsinspace")
				So(err, ShouldBeNil)
			})
			Convey("Chef should be able to prepare for sending a message", func() {
				chef.Nickname = "chef"
				err := chef.Connect()
				So(err, ShouldBeNil)
			})
			Convey("Beaker should be able to receive messages send by Chef", func() {
				go beaker.Listen()
				complete := make(chan common.Message, 1)
				go func() {
					m := <-beaker.Rcv
					complete <- m
				}()
				msg := new(common.Message)
				msg.Nickname = "chef"
				msg.Body = "dupdupdupdup"
				msg.Channel = "muppetsinspace"
				err := chef.SendMessage(msg)
				So(err, ShouldBeNil)

				select {
				case m := <-complete:
					So(m.Body, ShouldEqual, "dupdupdupdup")
					So(m.Nickname, ShouldEqual, "chef")
				case <-time.After(time.Second * 2):
					So(true, ShouldBeFalse)
				}
			})
		})

	})
	func() {
		res := common.MustGetRedis().Del(common.KeyWithPrefix(feeder.BOT_COUNT))
		if res.Err() != nil {
			fmt.Println(res.Err())
		}
		res = common.MustGetRedis().Del(common.KeyWithPrefix(common.REQ_CHANNELS_KEY))
		if res.Err() != nil {
			fmt.Println(res.Err())
		}
		feeder.Close()
	}()

}
