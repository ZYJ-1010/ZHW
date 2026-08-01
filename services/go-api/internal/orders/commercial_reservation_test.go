package orders

import "testing"

func TestCommercialPaymentTransitionReservation(t *testing.T) {
	if !CanTransitionCommercialPayment(PaymentCreated, PaymentPending) || !CanTransitionCommercialPayment(PaymentPending, PaymentEscrowed) || !CanTransitionCommercialPayment(PaymentEscrowed, PaymentSettlementPending) || !CanTransitionCommercialPayment(PaymentSettlementPending, PaymentSettled) {
		t.Fatal("expected the commercial settlement path to be reserved")
	}
	if CanTransitionCommercialPayment(PaymentCreated, PaymentSettled) || CanTransitionCommercialPayment(PaymentSettled, PaymentRefunded) {
		t.Fatal("terminal or skipped payment transitions must stay blocked")
	}
}
