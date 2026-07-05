package models

type OrderModel struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	ProductID  string  `json:"product_id"`
	StoreID    string  `json:"store_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"`
}

type ProductModel struct {
	ID              string  `json:"id"`
	UID             string  `json:"uid"`
	Name            string  `json:"name"`
	Price           float64 `json:"price"`
	Description     string  `json:"description"`
	PhotoURL        string  `json:"photoUrl"`
	Stock           int     `json:"stock"`
	Category        string  `json:"category"`
	IsQrisActive    int     `json:"isQrisActive"`
	IsPaypalActive  int     `json:"isPaypalActive"`
	ShippingJNE     int     `json:"shippingJNE"`
	ShippingJNT     int     `json:"shippingJNT"`
	ShippingSicepat int     `json:"shippingSicepat"`
	ShippingPos     int     `json:"shippingPos"`
	ShippingInstant int     `json:"shippingInstant"`
	ShippingPickup  int     `json:"shippingPickup"`
}

type PushPayload struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	URL    string `json:"url"`
}

type SubscriptionRequest struct {
	UID           string `json:"uid"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
}

type UserModel struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoURL"`
	PhoneNumber string `json:"phoneNumber"`
}
