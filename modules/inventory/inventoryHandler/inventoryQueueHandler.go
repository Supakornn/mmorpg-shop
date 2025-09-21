package inventoryHandler

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/queue"
)

type (
	InventoryQueueHandlerService interface {
		AddPlayerItem()
		RemovePlayerItem()
		RollbackAddPlayerItem()
		RollbackRemovePlayerItem()
	}

	inventoryQueueHandler struct {
		cfg              *config.Config
		inventoryUsecase inventoryUsecase.InventoryUsecaseService
	}
)

func NewInventoryQueueHandler(cfg *config.Config, inventoryUsecase inventoryUsecase.InventoryUsecaseService) InventoryQueueHandlerService {
	return &inventoryQueueHandler{cfg, inventoryUsecase}
}

func (h *inventoryQueueHandler) InventoryConsumer(pctx context.Context) (sarama.PartitionConsumer, error) {
	worker, err := queue.ConnectConsumer([]string{h.cfg.Kafka.Url}, h.cfg.Kafka.ApiKey, h.cfg.Kafka.Secret)
	if err != nil {
		return nil, errors.New("error: connect consumer failed")
	}

	offset, err := h.inventoryUsecase.GetOffset(pctx)
	if err != nil {
		return nil, errors.New("error: get offset failed")
	}

	consumer, err := worker.ConsumePartition("inventory", 0, offset)
	if err != nil {
		log.Printf("Error: Try to consume partition with offset: %v", err.Error())
		consumer, err = worker.ConsumePartition("inventory", 0, 0)
		if err != nil {
			log.Printf("Error: consume partition failed: %v", err.Error())
			return nil, errors.New("error: consume partition failed")
		}
	}

	return consumer, nil
}

func (h *inventoryQueueHandler) AddPlayerItem() {
	ctx := context.Background()

	consumer, err := h.InventoryConsumer(ctx)
	if err != nil {
		return
	}
	defer consumer.Close()

	log.Println("Add player item consumer started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-consumer.Errors():
			log.Printf("Error: add player item consumer failed: %v", err.Error())
			continue
		case msg := <-consumer.Messages():
			if string(msg.Key) == "buy" {
				h.inventoryUsecase.UpsertOffset(ctx, msg.Offset+1)
				req := new(inventory.UpdateInventoryReq)
				if err := queue.DecodeMessage(req, msg.Value); err != nil {
					continue
				}

				h.inventoryUsecase.AddPlayerItemRes(ctx, h.cfg, req)
				log.Printf("info: add player item: topic: %s, offset: %d, value: %s", msg.Topic, msg.Offset, string(msg.Value))
			}
		case <-sigChan:
			log.Println("Add player item consumer stopped")
			return
		}
	}
}

func (h *inventoryQueueHandler) RemovePlayerItem() {

}

func (h *inventoryQueueHandler) RollbackAddPlayerItem() {
	ctx := context.Background()

	consumer, err := h.InventoryConsumer(ctx)
	if err != nil {
		return
	}
	defer consumer.Close()

	log.Println("Rollback add player item consumer started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-consumer.Errors():
			log.Printf("Error: rollback add player item consumer failed: %v", err.Error())
			continue
		case msg := <-consumer.Messages():
			if string(msg.Key) == "radd" {
				h.inventoryUsecase.UpsertOffset(ctx, msg.Offset+1)
				req := new(inventory.RollbackInventoryReq)
				if err := queue.DecodeMessage(req, msg.Value); err != nil {
					continue
				}

				h.inventoryUsecase.RollbackAddPlayerItem(ctx, h.cfg, req)
				log.Printf("info: rollback add player item: topic: %s, offset: %d, value: %s", msg.Topic, msg.Offset, string(msg.Value))
			}
		case <-sigChan:
			log.Println("Rollback add player item consumer stopped")
			return
		}
	}
}

func (h *inventoryQueueHandler) RollbackRemovePlayerItem() {

}
