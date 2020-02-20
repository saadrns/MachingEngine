// Copyright 2019 GitBitEx.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"errors"
	"fmt"

	"github.com/gitbitex/gitbitex-spot/models"
	"github.com/gitbitex/gitbitex-spot/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
)

func PlaceLendingOrder(userId int64, productId string, autoRenew bool, side string,
	size, rate, duration decimal.Decimal) (*models.LendingOrder, error) {
	product, err := GetProductById(productId)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New(fmt.Sprintf("product not found: %v", productId))
	}

	// var holdCurrency string
	// var holdSize decimal.Decimal
	//	holdCurrency, holdSize = product.BaseCurrency, size
	fmt.Println("product.Id:", product.Id)
	order := &models.LendingOrder{
		UserId:            userId,
		ProductId:         product.Id,
		Size:              size,
		AutoRenew:         autoRenew,
		Duration:          duration,
		RemainingDuration: duration,
		Status:            "new",
		Funds:             decimal.Zero,
		Rate:              rate,
		Side:              "offer",
		Settled:           false,
	}

	// tx
	db, err := mysql.SharedStore().BeginTx()
	// if err != nil {
	// 	return nil, err
	// }
	// defer func() { _ = db.Rollback() }()

	// err = HoldBalance(db, userId, holdCurrency, holdSize, models.BillTypeTrade)
	// if err != nil {
	// 	return nil, err
	// }
	err = db.AddLendingOrder(order)
	if err != nil {
		return nil, err
	}

	return order, db.CommitTx()
}

// func UpdateOrderStatus(orderId int64, oldStatus, newStatus models.OrderStatus) (bool, error) {
// 	return mysql.SharedStore().UpdateOrderStatus(orderId, oldStatus, newStatus)
// }

// func ExecuteFill(orderId int64) error {
// 	// tx
// 	db, err := mysql.SharedStore().BeginTx()
// 	if err != nil {
// 		return err
// 	}
// 	defer func() { _ = db.Rollback() }()

// 	order, err := db.GetOrderByIdForUpdate(orderId)
// 	if err != nil {
// 		return err
// 	}
// 	if order == nil {
// 		return fmt.Errorf("order not found: %v", orderId)
// 	}
// 	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCancelled {
// 		return fmt.Errorf("order status invalid: %v %v", orderId, order.Status)
// 	}

// 	product, err := GetProductById(order.ProductId)
// 	if err != nil {
// 		return err
// 	}
// 	if product == nil {
// 		return fmt.Errorf("product not found: %v", order.ProductId)
// 	}

// 	fills, err := mysql.SharedStore().GetUnsettledFillsByOrderId(orderId)
// 	if err != nil {
// 		return err
// 	}
// 	if len(fills) == 0 {
// 		return nil
// 	}

// 	var bills []*models.Bill
// 	for _, fill := range fills {
// 		fill.Settled = true

// 		notes := fmt.Sprintf("%v-%v", fill.OrderId, fill.Id)

// 		if !fill.Done {
// 			executedValue := fill.Size.Mul(fill.Price)
// 			order.ExecutedValue = order.ExecutedValue.Add(executedValue)
// 			order.FilledSize = order.FilledSize.Add(fill.Size)

// 			if order.Side == models.SideBuy {
// 				// 买单，incr base
// 				bill, err := AddDelayBill(db, order.UserId, product.BaseCurrency, fill.Size, decimal.Zero,
// 					models.BillTypeTrade, notes)
// 				if err != nil {
// 					return err
// 				}
// 				bills = append(bills, bill)

// 				// 买单，decr quote
// 				bill, err = AddDelayBill(db, order.UserId, product.QuoteCurrency, decimal.Zero, executedValue.Neg(),
// 					models.BillTypeTrade, notes)
// 				if err != nil {
// 					return err
// 				}
// 				bills = append(bills, bill)

// 			} else {
// 				// 卖单，decr base
// 				bill, err := AddDelayBill(db, order.UserId, product.BaseCurrency, decimal.Zero, fill.Size.Neg(),
// 					models.BillTypeTrade, notes)
// 				if err != nil {
// 					return err
// 				}
// 				bills = append(bills, bill)

// 				// 卖单，incr quote
// 				bill, err = AddDelayBill(db, order.UserId, product.QuoteCurrency, executedValue, decimal.Zero,
// 					models.BillTypeTrade, notes)
// 				if err != nil {
// 					return err
// 				}
// 				bills = append(bills, bill)
// 			}

// 		} else {
// 			if fill.DoneReason == models.DoneReasonCancelled {
// 				order.Status = models.OrderStatusCancelled
// 			} else if fill.DoneReason == models.DoneReasonFilled {
// 				order.Status = models.OrderStatusFilled
// 			} else {
// 				log.Fatalf("unknown done reason: %v", fill.DoneReason)
// 			}

// 			if order.Side == models.SideBuy {
// 				// 如果是是买单，需要解冻剩余的funds
// 				remainingFunds := order.Funds.Sub(order.ExecutedValue)
// 				if remainingFunds.GreaterThan(decimal.Zero) {
// 					bill, err := AddDelayBill(db, order.UserId, product.QuoteCurrency, remainingFunds, remainingFunds.Neg(),
// 						models.BillTypeTrade, notes)
// 					if err != nil {
// 						return err
// 					}
// 					bills = append(bills, bill)
// 				}

// 			} else {
// 				// 如果是卖单，解冻剩余的size
// 				remainingSize := order.Size.Sub(order.FilledSize)
// 				if remainingSize.GreaterThan(decimal.Zero) {
// 					bill, err := AddDelayBill(db, order.UserId, product.BaseCurrency, remainingSize, remainingSize.Neg(),
// 						models.BillTypeTrade, notes)
// 					if err != nil {
// 						return err
// 					}
// 					bills = append(bills, bill)
// 				}
// 			}

// 			break
// 		}
// 	}

// 	err = db.UpdateOrder(order)
// 	if err != nil {
// 		return err
// 	}

// 	for _, fill := range fills {
// 		err = db.UpdateFill(fill)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return db.CommitTx()
// }

// func GetOrderById(orderId int64) (*models.Order, error) {
// 	return mysql.SharedStore().GetOrderById(orderId)
// }

// func GetOrderByClientOid(userId int64, clientOid string) (*models.Order, error) {
// 	return mysql.SharedStore().GetOrderByClientOid(userId, clientOid)
// }

// func GetOrdersByUserId(userId int64, statuses []models.OrderStatus, side *models.Side, productId string,
// 	beforeId, afterId int64, limit int) ([]*models.Order, error) {
// 	return mysql.SharedStore().GetOrdersByUserId(userId, statuses, side, productId, beforeId, afterId, limit)
// }

func PlaceMarginOrder(userId int64, productId string, types string,
	size, rate, price, leverage decimal.Decimal) (*models.MarginOrder, error) {
	product, err := GetProductById(productId)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New(fmt.Sprintf("product not found: %v", productId))
	}

	// var holdCurrency string
	// var holdSize decimal.Decimal
	//	holdCurrency, holdSize = product.BaseCurrency, size
	fmt.Println("product.Id:", product.Id)
	order := &models.MarginOrder{
		UserId:       userId,
		ProductId:    product.Id,
		Size:         size,
		Status:       "new",
		Funds:        decimal.Zero,
		Rate:         rate,
		Side:         "demand",
		Price:        price,
		Type:         types,
		Settled:      false,
		Category:     "limit",
		Leverage:     leverage,
		TriggerPrice: decimal.Zero,
	}

	// tx
	db, err := mysql.SharedStore().BeginTx()
	// if err != nil {
	// 	return nil, err
	// }
	// defer func() { _ = db.Rollback() }()

	// err = HoldBalance(db, userId, holdCurrency, holdSize, models.BillTypeTrade)
	// if err != nil {
	// 	return nil, err
	// }
	err = db.AddMarginOrder(order)
	if err != nil {
		return nil, err
	}

	return order, db.CommitTx()
}
