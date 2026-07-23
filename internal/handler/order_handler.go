package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunilsinghx/order-processing/internal/metrics"
	"github.com/sunilsinghx/order-processing/internal/repository"
	"github.com/sunilsinghx/order-processing/internal/service"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type CreateOrderRequest struct {
	CustomerName string  `json:"customer_name" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
}

func (h *OrderHandler) Create(c *gin.Context) {

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.CreateOrder(c.Request.Context(), req.CustomerName, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.OrdersCreated.Inc()

	// Return 202 Status Accepted
	c.JSON(http.StatusAccepted, order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order id is required",
		})
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "order not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch order",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}
