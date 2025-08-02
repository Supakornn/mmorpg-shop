package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentHandler"
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentRepository"
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentUsecase"
)

func (s *server) paymentService() {
	repo := paymentRepository.NewPaymentRepository(s.db)
	usecase := paymentUsecase.NewPaymentUsecase(repo)
	httpHandler := paymentHandler.NewPaymentHttpHandler(s.cfg, usecase)
	queueHandler := paymentHandler.NewPaymentQueueHandler(s.cfg, usecase)

	_ = httpHandler
	_ = queueHandler

	payment := s.app.Group("/payment_v1")

	// Health check
	_ = payment
}
