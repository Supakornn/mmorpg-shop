package playerHandler

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/queue"
)

type PlayerQueueHandlerService interface {
	DockedPlayerMoney()
	RollbackPlayerTransaction()
}

type playerQueueHandler struct {
	cfg           *config.Config
	playerUsecase playerUsecase.PlayerUsecaseService
}

func NewPlayerQueueHandler(cfg *config.Config, playerUsecase playerUsecase.PlayerUsecaseService) PlayerQueueHandlerService {
	return &playerQueueHandler{cfg, playerUsecase}
}

func (h *playerQueueHandler) PlayerConsumer(pctx context.Context) (sarama.PartitionConsumer, error) {
	worker, err := queue.ConnectConsumer([]string{h.cfg.Kafka.Url}, h.cfg.Kafka.ApiKey, h.cfg.Kafka.Secret)
	if err != nil {
		return nil, errors.New("error: connect consumer failed")
	}

	offset, err := h.playerUsecase.GetOffset(pctx)
	if err != nil {
		return nil, errors.New("error: get offset failed")
	}

	consumer, err := worker.ConsumePartition("player", 0, offset)
	if err != nil {
		log.Printf("Error: Try to consume partition with offset: %v", err.Error())
		consumer, err = worker.ConsumePartition("player", 0, 0)
		if err != nil {
			log.Printf("Error: consume partition failed: %v", err.Error())
			return nil, errors.New("error: consume partition failed")
		}
	}

	return consumer, nil
}

func (h *playerQueueHandler) DockedPlayerMoney() {
	ctx := context.Background()

	consumer, err := h.PlayerConsumer(ctx)
	if err != nil {
		return
	}
	defer consumer.Close()

	log.Println("Docked player money consumer started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-consumer.Errors():
			log.Printf("Error: docked player money consumer failed: %v", err.Error())
			continue
		case msg := <-consumer.Messages():
			if string(msg.Key) == "buy" {
				h.playerUsecase.UpsertOffset(ctx, msg.Offset+1)
				req := new(player.CreatePlayerTransactionReq)
				if err := queue.DecodeMessage(req, msg.Value); err != nil {
					continue
				}

				h.playerUsecase.DockedPlayerMoneyRes(ctx, h.cfg, req)
				log.Printf("info: docked player money: topic: %s, offset: %d, value: %s", msg.Topic, msg.Offset, string(msg.Value))
			}
		case <-sigChan:
			log.Println("Docked player money consumer stopped")
			return
		}
	}
}

func (h *playerQueueHandler) RollbackPlayerTransaction() {
	ctx := context.Background()

	consumer, err := h.PlayerConsumer(ctx)
	if err != nil {
		return
	}
	defer consumer.Close()

	log.Println("Rollback player transaction consumer started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-consumer.Errors():
			log.Printf("Error: rollback player transaction consumer failed: %v", err.Error())
			continue
		case msg := <-consumer.Messages():
			if string(msg.Key) == "rtransaction" {
				h.playerUsecase.UpsertOffset(ctx, msg.Offset+1)
				req := new(player.RollbackPlayerTransactionReq)
				if err := queue.DecodeMessage(req, msg.Value); err != nil {
					continue
				}

				h.playerUsecase.RollbackPlayerTransaction(ctx, req)
				log.Printf("info: rollback player transaction: topic: %s, offset: %d, value: %s", msg.Topic, msg.Offset, string(msg.Value))
			}
		case <-sigChan:
			log.Println("Rollback player transaction consumer stopped")
			return
		}
	}
}
