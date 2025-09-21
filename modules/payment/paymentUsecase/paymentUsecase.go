package paymentUsecase

import (
	"context"
	"errors"
	"log"

	"github.com/IBM/sarama"
	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	"github.com/Supakornn/mmorpg-shop/modules/item"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/payment"
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentRepository"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/Supakornn/mmorpg-shop/pkg/queue"
)

type (
	PaymentUsecaseService interface {
		FindeItemsInIds(pctx context.Context, grpcUrl string, req []*payment.ItemServiceReqDatum) error
		GetOffset(pctx context.Context) (int64, error)
		UpsertOffset(pctx context.Context, offset int64) error
		BuyItem(pctx context.Context, cfg *config.Config, playerId string, req *payment.ItemServiceReq) ([]*payment.PaymentTransferRes, error)
		SellItem(pctx context.Context, cfg *config.Config, playerId string, req *payment.ItemServiceReq) ([]*payment.PaymentTransferRes, error)
	}

	paymentUsecase struct {
		paymentRepository paymentRepository.PaymentRepositoryService
	}
)

func NewPaymentUsecase(paymentRepository paymentRepository.PaymentRepositoryService) PaymentUsecaseService {
	return &paymentUsecase{paymentRepository}
}

func (u *paymentUsecase) FindeItemsInIds(pctx context.Context, grpcUrl string, req []*payment.ItemServiceReqDatum) error {
	setIds := make(map[string]bool)
	for _, v := range req {
		if !setIds[v.ItemId] {
			setIds[v.ItemId] = true
		}
	}

	itemData, err := u.paymentRepository.FindItemsInIds(pctx, grpcUrl, &itemPb.FindItemsInIdsReq{
		Ids: func() []string {
			itemsIds := make([]string, 0)
			for k := range setIds {
				itemsIds = append(itemsIds, k)
			}
			return itemsIds
		}(),
	})
	if err != nil {
		log.Printf("Error: find items in ids failed: %v", err.Error())
		return errors.New("error: find items in ids failed")
	}

	itemMaps := make(map[string]*item.ItemShowCase)
	for _, data := range itemData.Items {
		itemMaps[data.Id] = &item.ItemShowCase{
			ItemId:   data.Id,
			Title:    data.Title,
			Price:    data.Price,
			ImageUrl: data.ImageUrl,
			Damage:   int(data.Damage),
		}
	}

	for i := range req {
		if _, ok := itemMaps[req[i].ItemId]; !ok {
			log.Printf("Error: item not found: %v", req[i].ItemId)
			return errors.New("error: item not found")
		}

		req[i].Price = itemMaps[req[i].ItemId].Price
	}

	return nil
}

func (u *paymentUsecase) GetOffset(pctx context.Context) (int64, error) {
	return u.paymentRepository.GetOffset(pctx)
}

func (u *paymentUsecase) UpsertOffset(pctx context.Context, offset int64) error {
	return u.paymentRepository.UpsertOffset(pctx, offset)
}

func (u *paymentUsecase) PaymentConsumer(pctx context.Context, cfg *config.Config) (sarama.PartitionConsumer, error) {
	worker, err := queue.ConnectConsumer([]string{cfg.Kafka.Url}, cfg.Kafka.ApiKey, cfg.Kafka.Secret)
	if err != nil {
		return nil, errors.New("error: connect consumer failed")
	}

	offset, err := u.GetOffset(pctx)
	if err != nil {
		return nil, errors.New("error: get offset failed")
	}

	consumer, err := worker.ConsumePartition("payment", 0, offset)
	if err != nil {
		log.Printf("Error: Try to consume partition with offset: %v", err.Error())
		consumer, err = worker.ConsumePartition("payment", 0, 0)
		if err != nil {
			log.Printf("Error: consume partition failed: %v", err.Error())
			return nil, errors.New("error: consume partition failed")
		}
	}

	return consumer, nil
}

func (u *paymentUsecase) TransactionConsumer(pctx context.Context, key string, cfg *config.Config, resCh chan<- *payment.PaymentTransferRes) {
	consumer, err := u.PaymentConsumer(pctx, cfg)
	if err != nil {
		resCh <- nil
		return
	}
	defer consumer.Close()

	log.Println("Transaction consumer started")

	select {
	case err := <-consumer.Errors():
		log.Printf("Error: transaction consumer failed: %v", err.Error())
		resCh <- nil
		return
	case msg := <-consumer.Messages():
		if string(msg.Key) == key {
			u.UpsertOffset(pctx, msg.Offset+1)
			req := new(payment.PaymentTransferRes)
			if err := queue.DecodeMessage(req, msg.Value); err != nil {
				resCh <- nil
				return
			}

			resCh <- req
			log.Printf("info: transaction: topic: %s, offset: %d, value: %s", msg.Topic, msg.Offset, string(msg.Value))
		}
	}
}

func (u *paymentUsecase) BuyItem(pctx context.Context, cfg *config.Config, playerId string, req *payment.ItemServiceReq) ([]*payment.PaymentTransferRes, error) {
	if err := u.FindeItemsInIds(pctx, cfg.Grpc.ItemUrl, req.Items); err != nil {
		log.Printf("Error: find items in ids failed: %v", err.Error())
		return nil, errors.New("error: find items in ids failed")
	}

	stage1 := make([]*payment.PaymentTransferRes, 0)
	for _, item := range req.Items {
		u.paymentRepository.DockedPlayerMoney(pctx, cfg, &player.CreatePlayerTransactionReq{
			PlayerId: playerId,
			Amount:   -item.Price,
		})

		resCh := make(chan *payment.PaymentTransferRes)

		go u.TransactionConsumer(pctx, "buy", cfg, resCh)

		res := <-resCh
		if res != nil {
			log.Printf("info: %v", res)
			stage1 = append(stage1, &payment.PaymentTransferRes{
				InventoryId:   "",
				TransactionId: res.TransactionId,
				PlayerId:      playerId,
				ItemId:        item.ItemId,
				Amount:        item.Price,
				Error:         res.Error,
			})
		}
	}

	for _, v := range stage1 {
		if v.Error != "" {
			for _, v2 := range stage1 {
				u.paymentRepository.RollbackTransaction(pctx, cfg, &player.RollbackPlayerTransactionReq{
					TransactionId: v2.TransactionId,
				})
			}

			return nil, errors.New(v.Error)
		}
	}

	stage2 := make([]*payment.PaymentTransferRes, 0)
	for _, s1 := range stage1 {
		u.paymentRepository.AddPlayerItem(pctx, cfg, &inventory.UpdateInventoryReq{
			PlayerId: playerId,
			ItemId:   s1.ItemId,
		})

		resCh := make(chan *payment.PaymentTransferRes)

		go u.TransactionConsumer(pctx, "buy", cfg, resCh)

		res := <-resCh
		if res != nil {
			log.Printf("info: %v", res)
			stage2 = append(stage2, &payment.PaymentTransferRes{
				InventoryId:   res.InventoryId,
				TransactionId: s1.TransactionId,
				PlayerId:      playerId,
				ItemId:        s1.ItemId,
				Amount:        s1.Amount,
				Error:         s1.Error,
			})
		}
	}

	for _, v := range stage2 {
		if v.Error != "" {
			for _, s2 := range stage2 {
				u.paymentRepository.RollbackAddPlayerItem(pctx, cfg, &inventory.RollbackInventoryReq{
					InventoryId: s2.InventoryId,
				})
			}

			for _, s2 := range stage2 {
				u.paymentRepository.RollbackTransaction(pctx, cfg, &player.RollbackPlayerTransactionReq{
					TransactionId: s2.TransactionId,
				})
			}

			return nil, errors.New(v.Error)
		}
	}

	return stage2, nil
}

func (u *paymentUsecase) SellItem(pctx context.Context, cfg *config.Config, playerId string, req *payment.ItemServiceReq) ([]*payment.PaymentTransferRes, error) {
	if err := u.FindeItemsInIds(pctx, cfg.Grpc.ItemUrl, req.Items); err != nil {
		log.Printf("Error: find items in ids failed: %v", err.Error())
		return nil, errors.New("error: find items in ids failed")
	}

	stage1 := make([]*payment.PaymentTransferRes, 0)
	for _, item := range req.Items {
		u.paymentRepository.RemovePlayerItem(pctx, cfg, &inventory.UpdateInventoryReq{
			PlayerId: playerId,
			ItemId:   item.ItemId,
		})

		resCh := make(chan *payment.PaymentTransferRes)

		go u.TransactionConsumer(pctx, "sell", cfg, resCh)

		res := <-resCh
		if res != nil {
			log.Printf("info: %v", res)
			stage1 = append(stage1, &payment.PaymentTransferRes{
				InventoryId:   "",
				TransactionId: "",
				PlayerId:      playerId,
				ItemId:        item.ItemId,
				Amount:        item.Price,
				Error:         res.Error,
			})
		}
	}

	for _, v := range stage1 {
		if v.Error != "" {
			for _, v2 := range stage1 {
				if v2.Error == "" {
					u.paymentRepository.RollbackRemovePlayerItem(pctx, cfg, &inventory.RollbackInventoryReq{
						PlayerId: playerId,
						ItemId:   v2.ItemId,
					})
				}
			}

			return nil, errors.New(v.Error)
		}
	}

	stage2 := make([]*payment.PaymentTransferRes, 0)
	for _, s1 := range stage1 {
		u.paymentRepository.AddPlayerMoney(pctx, cfg, &player.CreatePlayerTransactionReq{
			PlayerId: playerId,
			Amount:   s1.Amount * 0.8,
		})

		resCh := make(chan *payment.PaymentTransferRes)

		go u.TransactionConsumer(pctx, "sell", cfg, resCh)

		res := <-resCh
		if res != nil {
			log.Printf("info: %v", res)
			stage2 = append(stage2, &payment.PaymentTransferRes{
				InventoryId:   "",
				TransactionId: s1.TransactionId,
				PlayerId:      playerId,
				ItemId:        s1.ItemId,
				Amount:        s1.Amount,
				Error:         s1.Error,
			})
		}
	}

	for _, v := range stage2 {
		if v.Error != "" {
			for _, s2 := range stage2 {
				u.paymentRepository.RollbackTransaction(pctx, cfg, &player.RollbackPlayerTransactionReq{
					TransactionId: s2.TransactionId,
				})
			}

			for _, s2 := range stage2 {
				if s2.Error == "" {
					u.paymentRepository.RollbackRemovePlayerItem(pctx, cfg, &inventory.RollbackInventoryReq{
						InventoryId: s2.InventoryId,
					})
				}
			}

			return nil, errors.New(v.Error)
		}
	}

	return stage2, nil
}
