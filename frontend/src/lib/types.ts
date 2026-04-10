export type EventStatus = "pending" | "approved" | "rejected" | "settled";
export type UserRole = "user" | "moderator" | "admin";

export interface EventItem {
  id: string;
  creator_user_id: string;
  title: string;
  description: string;
  category: string;
  resolve_at: string;
  status: EventStatus;
  winner_outcome?: "yes" | "no";
  created_at: string;
}

export interface EventListResponse {
  items: EventItem[];
}

export interface ModerationQueueItem {
  task: {
    id: string;
    event_id: string;
    status: string;
    moderator_id?: string;
    reason?: string;
    created_at: string;
    reviewed_at?: string;
  };
  event: EventItem;
}

export interface ModerationEventsResponse {
  items: ModerationQueueItem[];
}

export interface CreateEventPayload {
  title: string;
  description: string;
  category: string;
  resolve_at: string;
}

export interface ApiErrorPayload {
  error: string;
  code?: string;
  details?: unknown;
}

export interface Wallet {
  user_id: string;
  balance_tokens: number;
  updated_at: string;
}

export interface WalletTransaction {
  id: string;
  user_id: string;
  type: string;
  amount_tokens: number;
  ref_type?: string;
  ref_id?: string;
  created_at: string;
}

export interface WalletTransactionsResponse {
  items: WalletTransaction[];
}

export type BetStatus = "open" | "won" | "lost" | "refunded";

export interface BetItem {
  id: string;
  user_id: string;
  event_id: string;
  outcome_code: "yes" | "no";
  stake: number;
  odds_at_bet: number;
  potential_payout: number;
  status: BetStatus;
  idempotency_key: string;
  placed_at: string;
  settled_at?: string;
}

export interface MyBetsResponse {
  items: BetItem[];
}

export interface PlaceBetPayload {
  event_id: string;
  outcome_code: "yes" | "no";
  stake: number;
}

export interface RegisterPayload {
  email: string;
  password: string;
}

export interface RegisterResponse {
  id: string;
  email: string;
  role: string;
  email_verified: boolean;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
}

export interface VerifyEmailResponse {
  id: string;
  email: string;
  email_verified: boolean;
}
