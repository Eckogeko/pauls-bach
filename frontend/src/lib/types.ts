export interface User {
  id: number;
  username: string;
  balance: number;
  is_admin: boolean;
  bingo: boolean;
  created_at: string;
}

export interface OutcomeOdds {
  outcome_id: number;
  label: string;
  odds: number;
  shares: number;
}

export interface Event {
  id: number;
  title: string;
  description: string;
  event_type: "binary" | "multi";
  status: "open" | "closed" | "resolved";
  winning_outcome_id?: number;
  created_at: string;
  resolved_at?: string;
  last_trade_at?: string;
  odds: OutcomeOdds[];
  bettors: Record<number, string[]>;
}

export interface EventDetail extends Event {
  user_positions?: UserPosition[];
}

export interface UserPosition {
  outcome_id: number;
  outcome_label: string;
  shares: number;
  avg_price: number;
}

export interface LeaderboardEntry {
  rank: number;
  username: string;
  balance: number;
  user_id: number;
}

export interface OddsSnapshot {
  id: number;
  event_id: number;
  outcome_id: number;
  odds: number;
  created_at: string;
}

export interface HistoryEntry {
  id: number;
  event_id: number;
  event_title: string;
  outcome_id: number;
  outcome_label: string;
  tx_type: "buy" | "sell" | "payout" | "bonus";
  shares: number;
  points: number;
  created_at: string;
}

export interface BingoEvent {
  id: number;
  title: string;
  rarity: "common" | "uncommon";
  resolved: boolean;
  created_at: string;
}

export interface BingoSquare {
  position: number;
  bingo_event_id?: number;
  custom_text?: string;
  resolved: boolean;
}

export interface BingoBoard {
  id: number;
  user_id: number;
  squares: BingoSquare[];
  created_at: string;
}

export interface PortfolioPosition {
  event_id: number;
  event_title: string;
  outcome_id: number;
  outcome_label: string;
  shares: number;
  avg_price: number;
  potential_payout: number;
}

export interface Portfolio {
  positions: PortfolioPosition[];
  total_invested: number;
  total_potential: number;
  active_markets: number;
}

export interface ActivityEntry {
  id: number;
  type: string;
  message: string;
  user_id: number;
  event_id: number;
  created_at: string;
}

export interface BingoWinner {
  id: number;
  user_id: number;
  username: string;
  board_id: number;
  line: string;
  created_at: string;
}
