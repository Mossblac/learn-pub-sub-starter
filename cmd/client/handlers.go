package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

//this is the "handler" input for SubscribeJson()

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(MoveChannel *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(MoveChannel, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{Attacker: move.Player, Defender: gs.GetPlayerSnap()})
			if err != nil {
				return pubsub.NackRequeue
			} else {
				return pubsub.Ack
			}
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(MoveChannel *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%v won a war against %v", winner, loser)
			glog := pubsub.CreateGameLog(time.Now(), msg, gs.GetUsername())
			err := pubsub.PublishGob(MoveChannel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), glog)
			if err != nil {
				return pubsub.NackRequeue
			} else {
				return pubsub.Ack
			}
		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser)
			glog := pubsub.CreateGameLog(time.Now(), msg, gs.GetUsername())
			err := pubsub.PublishGob(MoveChannel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), glog)
			if err != nil {
				return pubsub.NackRequeue
			} else {
				return pubsub.Ack
			}
		default:
			fmt.Println("Error Handling War")
			return pubsub.NackDiscard
		}
	}
}
