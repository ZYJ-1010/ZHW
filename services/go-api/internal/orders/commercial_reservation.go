package orders

// 二期商业化状态机预留。当前一期不会创建这些订单，也不会调用支付或分账。
// 保留为独立类型，避免未来将一对一/定制服务混入免费局 payment_orders。
type CommercialOrderType string

const (
	CommercialOrderTypeGameJoin      CommercialOrderType = "game_join"
	CommercialOrderTypeOneToOne      CommercialOrderType = "one_to_one"
	CommercialOrderTypeCustomService CommercialOrderType = "custom_service"
)

type DistributionMethod string

const (
	DistributionNone              DistributionMethod = "none"
	DistributionEqualSplit        DistributionMethod = "equal_split"
	DistributionVoluntary         DistributionMethod = "voluntary"
	DistributionContributionBased DistributionMethod = "contribution_based"
)

type CommercialPaymentStatus string

const (
	PaymentCreated           CommercialPaymentStatus = "created"
	PaymentPending           CommercialPaymentStatus = "pending_payment"
	PaymentEscrowed          CommercialPaymentStatus = "escrowed"
	PaymentSettlementPending CommercialPaymentStatus = "settlement_pending"
	PaymentSettled           CommercialPaymentStatus = "settled"
	PaymentRefundPending     CommercialPaymentStatus = "refund_pending"
	PaymentRefunded          CommercialPaymentStatus = "refunded"
	PaymentClosed            CommercialPaymentStatus = "closed"
)

func CanTransitionCommercialPayment(from CommercialPaymentStatus, to CommercialPaymentStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case PaymentCreated:
		return to == PaymentPending || to == PaymentClosed
	case PaymentPending:
		return to == PaymentEscrowed || to == PaymentClosed
	case PaymentEscrowed:
		return to == PaymentSettlementPending || to == PaymentRefundPending
	case PaymentSettlementPending:
		return to == PaymentSettled || to == PaymentRefundPending
	case PaymentRefundPending:
		return to == PaymentRefunded
	default:
		return false
	}
}

// CommercialOrderReservation is intentionally transport-neutral. The second
// phase repository/service will persist it in commercial_service_orders and
// submit the payment only after compliance and merchant agreements are ready.
type CommercialOrderReservation struct {
	OrderType          CommercialOrderType     `json:"orderType"`
	GameID             int64                   `json:"gameId,omitempty"`
	BuyerUserID        int64                   `json:"buyerUserId"`
	ProviderUserID     int64                   `json:"providerUserId"`
	AmountCent         int64                   `json:"amountCent"`
	PaymentStatus      CommercialPaymentStatus `json:"paymentStatus"`
	DistributionMethod DistributionMethod      `json:"distributionMethod"`
	PlatformFeeBPS     int                     `json:"platformFeeBps"`
}
