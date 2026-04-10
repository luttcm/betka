import {
  ApiErrorPayload,
  BetItem,
  CreateEventPayload,
  EventOdds,
  EventItem,
  EventListResponse,
  LoginPayload,
  LoginResponse,
  MyBetsResponse,
  ModerationEventsResponse,
  ModerationQueueItem,
  PlaceBetPayload,
  RequestSettlementPayload,
  RegisterPayload,
  RegisterResponse,
  VerifyEmailResponse,
  Wallet,
  WalletTransaction,
  WalletTransactionsResponse,
} from "@/lib/types";

const API_BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") ?? "/api").replace(
  /\/$/,
  "",
);

export class ApiError extends Error {
  status: number;
  code?: string;
  details?: unknown;

  constructor(message: string, status: number, payload?: ApiErrorPayload | null) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = payload?.code;
    this.details = payload?.details;
  }
}

function isApiErrorPayload(data: unknown): data is ApiErrorPayload {
  return Boolean(data && typeof data === "object" && "error" in data && typeof data.error === "string");
}

function withAuth(token: string): Record<string, string> {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
}

async function parseResponse<T>(response: Response): Promise<T> {
  const data = await response.json().catch(() => null);

  if (!response.ok) {
    const message = isApiErrorPayload(data) ? data.error : `request failed with status ${response.status}`;

    throw new ApiError(message, response.status, isApiErrorPayload(data) ? data : null);
  }

  return data as T;
}

export async function getEvents(): Promise<EventItem[]> {
  const response = await fetch(`${API_BASE_URL}/v1/events`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  const payload = await parseResponse<EventListResponse>(response);
  return payload.items;
}

export async function getEventById(eventId: string): Promise<EventItem> {
  const response = await fetch(`${API_BASE_URL}/v1/events/${eventId}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  return parseResponse<EventItem>(response);
}

export async function getEventOdds(eventId: string): Promise<EventOdds> {
  const response = await fetch(`${API_BASE_URL}/v1/events/${eventId}/odds`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  return parseResponse<EventOdds>(response);
}

export async function createEvent(payload: CreateEventPayload, token: string): Promise<EventItem> {
  const response = await fetch(`${API_BASE_URL}/v1/events`, {
    method: "POST",
    headers: withAuth(token),
    body: JSON.stringify(payload),
  });

  return parseResponse<EventItem>(response);
}

export async function requestSettlement(eventId: string, payload: RequestSettlementPayload, token: string): Promise<EventItem> {
  const response = await fetch(`${API_BASE_URL}/v1/events/${eventId}/request-settlement`, {
    method: "POST",
    headers: withAuth(token),
    body: JSON.stringify(payload),
  });

  return parseResponse<EventItem>(response);
}

export async function getAdminSettlementRequests(token: string): Promise<EventItem[]> {
  const response = await fetch(`${API_BASE_URL}/v1/admin/events/settlement-requests`, {
    method: "GET",
    headers: withAuth(token),
    cache: "no-store",
  });

  const payload = await parseResponse<EventListResponse>(response);
  return payload.items;
}

export async function settleAdminEvent(eventId: string, winnerOutcome: "yes" | "no", token: string): Promise<{
  event: EventItem;
  settled_bets: BetItem[];
}> {
  const response = await fetch(`${API_BASE_URL}/v1/admin/events/${eventId}/settle`, {
    method: "POST",
    headers: withAuth(token),
    body: JSON.stringify({ winner_outcome: winnerOutcome }),
  });

  return parseResponse<{ event: EventItem; settled_bets: BetItem[] }>(response);
}

export async function register(payload: RegisterPayload): Promise<RegisterResponse> {
  const response = await fetch(`${API_BASE_URL}/v1/auth/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  return parseResponse<RegisterResponse>(response);
}

export async function login(payload: LoginPayload): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/v1/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  return parseResponse<LoginResponse>(response);
}

export async function verifyEmail(token: string): Promise<VerifyEmailResponse> {
  const params = new URLSearchParams({ token });
  const response = await fetch(`${API_BASE_URL}/v1/auth/verify-email?${params.toString()}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  return parseResponse<VerifyEmailResponse>(response);
}

export async function getModerationEvents(token: string): Promise<ModerationQueueItem[]> {
  const response = await fetch(`${API_BASE_URL}/v1/moderation/events`, {
    method: "GET",
    headers: withAuth(token),
    cache: "no-store",
  });

  const payload = await parseResponse<ModerationEventsResponse>(response);
  return payload.items;
}

export async function approveModerationEvent(eventId: string, token: string): Promise<EventItem> {
  const response = await fetch(`${API_BASE_URL}/v1/moderation/events/${eventId}/approve`, {
    method: "POST",
    headers: withAuth(token),
  });

  return parseResponse<EventItem>(response);
}

export async function rejectModerationEvent(eventId: string, reason: string, token: string): Promise<EventItem> {
  const response = await fetch(`${API_BASE_URL}/v1/moderation/events/${eventId}/reject`, {
    method: "POST",
    headers: withAuth(token),
    body: JSON.stringify({ reason }),
  });

  return parseResponse<EventItem>(response);
}

export async function getWallet(token: string): Promise<Wallet> {
  const response = await fetch(`${API_BASE_URL}/v1/wallet`, {
    method: "GET",
    headers: withAuth(token),
    cache: "no-store",
  });

  return parseResponse<Wallet>(response);
}

export async function getWalletTransactions(token: string): Promise<WalletTransaction[]> {
  const response = await fetch(`${API_BASE_URL}/v1/wallet/transactions`, {
    method: "GET",
    headers: withAuth(token),
    cache: "no-store",
  });

  const payload = await parseResponse<WalletTransactionsResponse>(response);
  return payload.items;
}

export async function getMyBets(token: string): Promise<BetItem[]> {
  const response = await fetch(`${API_BASE_URL}/v1/bets/my`, {
    method: "GET",
    headers: withAuth(token),
    cache: "no-store",
  });

  const payload = await parseResponse<MyBetsResponse>(response);
  return payload.items;
}

export async function placeBet(payload: PlaceBetPayload, token: string, idempotencyKey: string): Promise<BetItem> {
  const response = await fetch(`${API_BASE_URL}/v1/bets`, {
    method: "POST",
    headers: {
      ...withAuth(token),
      "Idempotency-Key": idempotencyKey,
    },
    body: JSON.stringify(payload),
  });

  return parseResponse<BetItem>(response);
}
