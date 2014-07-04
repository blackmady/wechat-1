// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat for the canonical source repository
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package client

import (
	"github.com/chanxuehong/wechat/merchant/order"
)

// 根据订单id获取订单详情
func (c *Client) MerchantOrderGetById(orderId string) (*order.Order, error) {
	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := merchantOrderGetByIdURL(token)

	var request = struct {
		OrderId string `json:"order_id"`
	}{
		OrderId: orderId,
	}

	var result struct {
		Error
		Order order.Order `json:"order"`
	}
	if err = c.postJSON(_url, request, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}

	return &result.Order, nil
}

// 根据订单状态/创建时间获取订单详情
//  @status:    订单状态(不带该字段 == 0 -全部状态, 2-待发货, 3-已发货, 5-已完成, 8-维权中)
//  @beginTime: 订单创建时间起始时间(不带该字段 ==0 则不按照时间做筛选)
//  @endTime:   订单创建时间终止时间(不带该字段 ==0 则不按照时间做筛选)
func (c *Client) MerchantOrderGetByFilter(status int, beginTime, endTime int64) ([]order.Order, error) {
	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := merchantOrderGetByFilterURL(token)

	var request = struct {
		Status    int   `json:"status,omitempty"`
		BeginTime int64 `json:"begintime,omitempty"`
		EndTime   int64 `json:"endtime,omitempty"`
	}{
		Status:    status,
		BeginTime: beginTime,
		EndTime:   endTime,
	}

	var result struct {
		Error
		OrderList []order.Order `json:"order_list"`
	}
	result.OrderList = make([]order.Order, 0, 256)
	if err = c.postJSON(_url, &request, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}

	return result.OrderList, nil
}

// 设置订单发货信息.
//  @orderId:         订单ID;
//  @deliveryCompany: 物流公司ID(参考《物流公司ID》)
//  @deliveryTrackNo: 运单ID
func (c *Client) MerchantOrderSetDelivery(orderId, deliveryCompany, deliveryTrackNo string) error {
	token, err := c.Token()
	if err != nil {
		return err
	}
	_url := merchantOrderSetDeliveryURL(token)

	var request = struct {
		OrderId         string `json:"order_id"`
		DeliveryCompany string `json:"delivery_company"`
		DeliveryTrackNo string `json:"delivery_track_no"`
	}{
		OrderId:         orderId,
		DeliveryCompany: deliveryCompany,
		DeliveryTrackNo: deliveryTrackNo,
	}

	var result Error
	if err = c.postJSON(_url, &request, &result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return &result
	}

	return nil
}

// 关闭订单
func (c *Client) MerchantOrderClose(orderId string) error {
	token, err := c.Token()
	if err != nil {
		return err
	}
	_url := merchantOrderCloseURL(token)

	var request = struct {
		OrderId string `json:"order_id"`
	}{
		OrderId: orderId,
	}

	var result Error
	if err = c.postJSON(_url, request, &result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return &result
	}

	return nil
}
