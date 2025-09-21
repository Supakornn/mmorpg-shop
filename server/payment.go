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

	payment := s.app.Group("/payment_v1")

	// Health check
	payment.GET("", s.healthCheckService)
	payment.POST("/payment/buy", httpHandler.BuyItem, s.mid.JwtAuthorization)
	payment.POST("/payment/sell", httpHandler.SellItem, s.mid.JwtAuthorization)
}
