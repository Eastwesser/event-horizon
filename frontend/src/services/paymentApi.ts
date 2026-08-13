import api from './api';

export interface SubscriptionStatus {
  active: boolean;
  plan: string;
  status: string;
  expires_at_unix: number;
  amount_rub: number;
}

export interface CheckoutResponse {
  payment_id: string;
  checkout_url: string;
  amount_rub: number;
  plan: string;
  status: string;
}

export const paymentApi = {
  getSubscription: async (): Promise<SubscriptionStatus> => {
    const { data } = await api.get('/api/payment/subscription');
    return data;
  },
  createCheckout: async (plan: 'present' | 'future'): Promise<CheckoutResponse> => {
    const { data } = await api.post('/api/payment/checkout', { plan });
    return data;
  },
  canPurchaseMerch: async (): Promise<{ allowed: boolean; reason: string }> => {
    const { data } = await api.get('/api/payment/can-purchase-merch');
    return data;
  },
};
