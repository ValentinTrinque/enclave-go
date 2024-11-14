package apiclient

import (
	"fmt"

	"github.com/Enclave-Markets/enclave-go/models"
)

func (client *ApiClient) AddSpotOrder(req models.AddOrderReq) (*models.GenericResponse[models.ApiOrder], error) {
	path := models.V1SpotOrdersPath

	res, err := NewHttpJsonClient[models.AddOrderReq, models.GenericResponse[models.ApiOrder]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("POST", path, req)).Post(req)

	if err != nil {
		return res, fmt.Errorf("error with http req in spot add order: %w", err)
	}
	if !res.Success {
		return res, fmt.Errorf("error in spot add order %v: %v", req, res.Error)
	}

	return res, err
}

func (client *ApiClient) GetSpotDepthBook(market models.Market) (*models.GenericResponse[models.BookSnapshot], error) {
	path := models.V1SpotDepthPath + "?market=" + string(market)

	res, err := NewHttpJsonClient[any, models.GenericResponse[models.BookSnapshot]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("GET", path, nil)).Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error in http req Spot get depth book: %w", err)
	}
	if !res.Success {
		return res, fmt.Errorf("bad request Spot get depth book: %v", res.Error)
	}

	return res, nil
}

func (client *ApiClient) GetSpotOrder(orderId models.OrderID) (*models.GenericResponse[models.ApiOrder], error) {
	path := models.V1SpotOrdersPath + "/" + string(orderId)

	res, err := NewHttpJsonClient[any, models.GenericResponse[models.ApiOrder]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("GET", path, nil)).Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error in http req spot get order: %w", err)
	}
	if !res.Success {
		return res, fmt.Errorf("bad request spot get order %s: %v", orderId, res.Error)
	}

	return res, nil
}

func (client *ApiClient) CancelAllSpotOrders() error {
	path := models.V1SpotOrdersPath

	res, err := NewHttpJsonClient[any, models.GenericResponse[any]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("DELETE", path, nil)).Delete(nil)

	if err != nil {
		return fmt.Errorf("error in http req spot delete all orders: %w", err)
	}
	if !res.Success {
		return fmt.Errorf("bad request spot delete all orders: %v", res.Error)
	}

	return nil
}

func (client *ApiClient) CancelSpotOrder(orderId models.OrderID) (*models.GenericResponse[any], error) {
	path := models.V1SpotOrdersPath + "/" + string(orderId)

	res, err := NewHttpJsonClient[any, models.GenericResponse[any]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("DELETE", path, nil)).Delete(nil)

	if err != nil {
		return res, fmt.Errorf("error in http req spot delete order: %w", err)
	}
	if !res.Success {
		return res, fmt.Errorf("bad request spot delete order %s: %v", orderId, res.Error)
	}

	return res, nil
}

func (client *ApiClient) CancelSpotOrderByClientID(clientOrderId models.OrderID) (*models.GenericResponse[any], error) {
	path := models.V1SpotOrdersPath + "/" + models.V1SpotClientOrderIDPrefix + string(clientOrderId)

	res, err := NewHttpJsonClient[any, models.GenericResponse[any]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("DELETE", path, nil)).Delete(nil)

	if err != nil {
		return res, fmt.Errorf("error in http req spot delete order by client id: %w", err)
	}
	if !res.Success {
		return res, fmt.Errorf("bad request spot delete order by client id %s: %v", clientOrderId, res.Error)
	}

	return res, nil
}

func (client *ApiClient) GetSpotFills(params models.FillParams) (*models.V1PageRes[models.ApiFill], error) {
	path := models.V1SpotFillsPath
	path += params.GetFillPathParams()

	res, err := NewHttpJsonClient[any, models.V1PageRes[models.ApiFill]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("GET", path, nil)).Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error in http req spot get fills: %w", err)
	}

	return res, err
}

func (client *ApiClient) GetSpotFillsByOrderID(orderID models.OrderID) (*models.GenericResponse[[]models.ApiFill], error) {
	path := models.V1SpotOrdersPath + "/" + string(orderID) + "/fills"

	res, err := NewHttpJsonClient[any, models.GenericResponse[[]models.ApiFill]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("GET", path, nil)).Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error in http req spot get fill by order ID: %w", err)
	}

	if !res.Success {
		return res, fmt.Errorf("bad request spot fill by order id %s: %v", orderID, res.Error)
	}

	return res, err
}

func (client *ApiClient) GetSpotFillsByClientOrderID(orderID models.OrderID) (*models.GenericResponse[[]models.ApiFill], error) {
	path := models.V1SpotOrdersPath + "/" + models.V1SpotClientOrderIDPrefix + string(orderID) + "/fills"

	res, err := NewHttpJsonClient[any, models.GenericResponse[[]models.ApiFill]](
		client.ApiEndpoint + path).SetHeaders(client.getHeaders("GET", path, nil)).Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error in http req spot get fill by client order ID: %w", err)
	}

	if !res.Success {
		return res, fmt.Errorf("bad request spot fill by client order id %s: %v", orderID, res.Error)
	}

	return res, err
}
