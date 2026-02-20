const BASE = "";

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem("token");
  let res: Response;
  try {
    res = await fetch(BASE + path, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options?.headers,
      },
    });
  } catch {
    throw new ApiError(0, "Network error — check your connection");
  }
  if (res.status === 401) {
    localStorage.removeItem("token");
    window.location.href = "/login";
    throw new ApiError(401, "Session expired — please log in again");
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }));
    throw new ApiError(res.status, body.error || "Unknown error");
  }
  return res.json();
}

export { api, ApiError };

// Auth
export const login = (username: string, pin: string) =>
  api<{ token: string; user: import("./types").User }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, pin }),
  });

export const register = (username: string, pin: string) =>
  api<{ token: string; user: import("./types").User }>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({ username, pin }),
  });

export const getMe = () =>
  api<import("./types").User>("/api/auth/me");

// Events
export const getEvents = () =>
  api<import("./types").Event[]>("/api/events");

export const getEvent = (id: number) =>
  api<import("./types").EventDetail>(`/api/events/${id}`);

export const getOddsHistory = (eventId: number) =>
  api<import("./types").OddsSnapshot[]>(`/api/events/${eventId}/odds-history`);

// Trading
export const buyShares = (eventId: number, outcomeId: number, amount: number) =>
  api<{ message: string; odds: import("./types").OutcomeOdds[]; balance: number }>(
    `/api/events/${eventId}/buy`,
    { method: "POST", body: JSON.stringify({ outcome_id: outcomeId, amount }) }
  );

export const sellShares = (eventId: number, outcomeId: number, shares: number) =>
  api<{ message: string; points_back: number; odds: import("./types").OutcomeOdds[]; balance: number }>(
    `/api/events/${eventId}/sell`,
    { method: "POST", body: JSON.stringify({ outcome_id: outcomeId, shares }) }
  );

// User event creation
export const createUserEvent = (title: string, description: string, eventType: string, outcomes: string[]) =>
  api<{ event: import("./types").Event }>("/api/events", {
    method: "POST",
    body: JSON.stringify({ title, description, event_type: eventType, outcomes }),
  });

// Admin
export const createEvent = (title: string, description: string, eventType: string, outcomes: string[]) =>
  api<{ event: import("./types").Event }>("/api/admin/events", {
    method: "POST",
    body: JSON.stringify({ title, description, event_type: eventType, outcomes }),
  });

export const updateEvent = (eventId: number, data: { title: string; description: string; outcomes: { id?: number; label: string }[] }) =>
  api<{ message: string }>(`/api/admin/events/${eventId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteEvent = (eventId: number) =>
  api<{ message: string }>(`/api/admin/events/${eventId}`, {
    method: "DELETE",
  });

export const resolveEvent = (eventId: number, winningOutcomeId: number) =>
  api<{ message: string }>(`/api/admin/events/${eventId}/resolve`, {
    method: "POST",
    body: JSON.stringify({ winning_outcome_id: winningOutcomeId }),
  });

export const unresolveEvent = (eventId: number) =>
  api<{ message: string }>(`/api/admin/events/${eventId}/unresolve`, {
    method: "POST",
  });

// Leaderboard
export const getLeaderboard = () =>
  api<import("./types").LeaderboardEntry[]>("/api/leaderboard");

// History
export const getHistory = (userId: number) =>
  api<import("./types").HistoryEntry[]>(`/api/users/${userId}/history`);

// Admin - Users
export const getAdminUsers = () =>
  api<{ id: number; username: string; is_admin: boolean; bingo: boolean; balance: number }[]>("/api/admin/users");

export const setUserBingo = (userId: number, bingo: boolean) =>
  api<{ message: string; bingo: boolean }>(`/api/admin/users/${userId}/bingo`, {
    method: "POST",
    body: JSON.stringify({ bingo }),
  });

export const setUserBalance = (userId: number, balance: number) =>
  api<{ message: string; balance: number }>(`/api/admin/users/${userId}/balance`, {
    method: "POST",
    body: JSON.stringify({ balance }),
  });

export const resetUserBingo = (userId: number) =>
  api<{ message: string }>(`/api/admin/users/${userId}/reset-bingo`, {
    method: "POST",
  });

// Bingo
export const getBingoEvents = () =>
  api<import("./types").BingoEvent[]>("/api/bingo/events");

export const getBingoBoard = () =>
  api<{ board: import("./types").BingoBoard; winners: import("./types").BingoWinner[] }>("/api/bingo/board");

export const createBingoBoard = (squares: import("./types").BingoSquare[]) =>
  api<import("./types").BingoBoard>("/api/bingo/board", {
    method: "POST",
    body: JSON.stringify({ squares }),
  });

export const getBingoWinners = () =>
  api<import("./types").BingoWinner[]>("/api/bingo/winners");

export const getAllBingoBoards = () =>
  api<(import("./types").BingoBoard & { username: string; winners: import("./types").BingoWinner[] })[]>("/api/bingo/boards");

// Portfolio
export const getPortfolio = () =>
  api<import("./types").Portfolio>("/api/portfolio");

// Activity
export const getActivity = () =>
  api<import("./types").ActivityEntry[]>("/api/activity");

// Bingo Admin
export const createBingoEvent = (title: string, rarity: string = "common") =>
  api<import("./types").BingoEvent>("/api/admin/bingo/events", {
    method: "POST",
    body: JSON.stringify({ title, rarity }),
  });

export const updateBingoEvent = (eventId: number, data: { title: string; rarity: string }) =>
  api<import("./types").BingoEvent>(`/api/admin/bingo/events/${eventId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const resolveBingoEvent = (eventId: number) =>
  api<{ message: string }>(`/api/admin/bingo/events/${eventId}/resolve`, {
    method: "POST",
  });

export const unresolveBingoEvent = (eventId: number) =>
  api<{ message: string }>(`/api/admin/bingo/events/${eventId}/unresolve`, {
    method: "POST",
  });

