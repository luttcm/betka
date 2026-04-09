import {
  CreateEventPayload,
  EventItem,
  EventListResponse,
  LoginPayload,
  LoginResponse,
  RegisterPayload,
  RegisterResponse,
  VerifyEmailResponse,
} from "@/lib/types";

const API_BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") ?? "/api").replace(
  /\/$/,
  "",
);

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function parseResponse<T>(response: Response): Promise<T> {
  const data = await response.json().catch(() => null);

  if (!response.ok) {
    const message =
      data && typeof data === "object" && "error" in data && typeof data.error === "string"
        ? data.error
        : `request failed with status ${response.status}`;

    throw new ApiError(message, response.status);
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

export async function createEvent(payload: CreateEventPayload, token: string): Promise<EventItem> {
  const response = await fetch(`${API_BASE_URL}/v1/events`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  });

  return parseResponse<EventItem>(response);
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
