import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Auth Token Management
export function setAuthToken(token: string | null) {
  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`
  } else {
    delete api.defaults.headers.common['Authorization']
  }
}

// Types
export interface Player {
  replay_id: number
  player_id: number
  player_slot: number
  name: string
  race: string
  result: string
  apm: number
  spending_quotient: number
  is_human: boolean
}

export interface Replay {
  id: number
  hash: string
  filename: string
  map: string
  duration: number
  game_version: string
  played_at: string
  uploaded_at: string
  players: Player[]
}

export interface SupplyPoint {
  time: number
  supply_used: number
  supply_max: number
  is_blocked: boolean
}

export interface SupplyBlock {
  start_time: number
  end_time: number
  duration: number
  severity: string
  supply_used: number
  supply_max: number
}

export interface SupplyAnalysis {
  total_block_time: number
  block_percentage: number
  blocks: SupplyBlock[]
  supply_timeline: SupplyPoint[]
}

export interface ResourcePoint {
  time: number
  minerals: number
  gas: number
  income: {
    minerals: number
    gas: number
  }
}

export interface SpendingAnalysis {
  spending_quotient: number
  rating: string
  average_unspent: {
    minerals: number
    gas: number
  }
  average_income: {
    minerals: number
    gas: number
  }
  resource_timeline: ResourcePoint[]
}

export interface APMPoint {
  time: number
  apm: number
}

export interface APMAnalysis {
  average_apm: number
  peak_apm: number
  eapm: number
  apm_timeline: APMPoint[]
}

export interface BuildOrderItem {
  time: number
  supply: number
  action: string
  unit_or_building: string
}

export interface ArmyPoint {
  time: number
  value: number
  unit_count: number
}

export interface ArmyAnalysis {
  peak_army_value: number
  army_timeline: ArmyPoint[]
  unit_composition: {
    unit_type: string
    count: number
    value: number
  }[]
}

export interface Suggestion {
  priority: string
  category: string
  title: string
  description: string
  timestamp?: number
  target_value?: string
}

export interface AnalysisData {
  supply_analysis?: SupplyAnalysis
  spending_analysis?: SpendingAnalysis
  apm_analysis?: APMAnalysis
  build_order?: BuildOrderItem[]
  inject_analysis?: {
    efficiency: number
    total_injects: number
    missed_injects: number
  }
  army_analysis?: ArmyAnalysis
  suggestions: Suggestion[]
}

export interface ReplayAnalysis {
  replay: Replay
  analyses: Record<number, AnalysisData>
}

export interface TrendData {
  metric: string
  trend: string
  change: number
}

// Strategic Analysis Types
export interface MetricComparison {
  metric: string
  player_value: number
  enemy_value: number
  is_worse: boolean
}

export interface SupplyBlockSummary {
  time: number
  duration: number
  severity: string
}

export interface CriticalMoment {
  time: number
  player_loss: number
  enemy_loss: number
  assessment: string
  is_positive: boolean
}

export interface IdentifiedProblem {
  title: string
  description: string
  priority: string
}

export interface MatchupTips {
  opening: string[]
  mid_game: string[]
  timing: string[]
  late_game: string[]
}

export interface ImprovementStep {
  category: string
  title: string
  description: string
}

export interface StrategicAnalysis {
  winner: string
  loser: string
  winner_race: string
  loser_race: string
  matchup: string
  metrics_comparison: MetricComparison[]
  supply_blocks: SupplyBlockSummary[]
  critical_moments: CriticalMoment[]
  problems: IdentifiedProblem[]
  matchup_tips: MatchupTips
  improvement_steps: ImprovementStep[]
  summary: string
}

export interface StrategicAnalysisResponse {
  replay: Replay
  analysis: StrategicAnalysis
}

// API Functions
export interface UploadResponse {
  replay_id: number
  replay: Replay
  message: string
  needs_player_selection?: boolean
}

export async function uploadReplay(file: File): Promise<UploadResponse> {
  const formData = new FormData()
  formData.append('replay', file)

  const response = await api.post('/replays/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return response.data
}

export async function claimReplay(replayId: number, playerId: number): Promise<{ message: string; player_name: string }> {
  const response = await api.post(`/replays/${replayId}/claim`, { player_id: playerId })
  return response.data
}

export async function deleteReplay(replayId: number): Promise<void> {
  await api.delete(`/replays/${replayId}`)
}

export async function listReplays(limit = 20, offset = 0): Promise<{ replays: Replay[]; total: number }> {
  const response = await api.get('/replays', {
    params: { limit, offset },
  })
  return response.data
}

export async function getReplay(id: number): Promise<Replay> {
  const response = await api.get(`/replays/${id}`)
  return response.data
}

export async function getReplayAnalysis(id: number): Promise<ReplayAnalysis> {
  const response = await api.get(`/replays/${id}/analysis`)
  return response.data
}

export async function getTrends(playerId: number, limit = 20): Promise<{ trends: Record<string, TrendData> }> {
  const response = await api.get('/stats/trends', {
    params: { player_id: playerId, limit },
  })
  return response.data
}

export async function getStrategicAnalysis(id: number): Promise<StrategicAnalysisResponse> {
  const response = await api.get(`/replays/${id}/strategic`)
  return response.data
}

// ============== Auth Types ==============

export interface User {
  id: number
  email: string
  sc2_player_name: string
  created_at: string
  last_login?: string
}

export interface AuthResponse {
  token: string
  user: User
}

// ============== Mentor Types ==============

export interface Goal {
  id: number
  user_id: number
  goal_type: string
  metric_name: string
  target_value: number
  comparison: string
  current_value: number
  status: string
  created_at: string
  deadline: string
}

export interface GoalTemplate {
  name: string
  goal_type: string
  metric_name: string
  comparison: string
  beginner: number
  advanced: number
  description: string
}

export interface DailyProgress {
  id: number
  user_id: number
  date: string
  games_played: number
  wins: number
  losses: number
  avg_apm: number
  avg_spending_quotient: number
  avg_supply_block_pct: number
  total_play_time: number
}

export interface WeekStats {
  games_played: number
  wins: number
  losses: number
  win_rate: number
  avg_apm: number
  avg_sq: number
  avg_supply_block: number
  total_play_time: number
  apm_change: number
  sq_change: number
  win_rate_change: number
  supply_block_change: number
}

export interface RecentGame {
  replay_id: number
  map: string
  result: string
  race: string
  enemy_race: string
  apm: number
  sq: number
  duration: number
  played_at: string
}

export interface CoachingFocus {
  id: number
  user_id: number
  focus_area: string
  description: string
  started_at: string
  active: boolean
}

export interface WeeklyReport {
  id: number
  user_id: number
  week_start: string
  week_end: string
  total_games: number
  wins: number
  losses: number
  win_rate: number
  avg_apm: number
  avg_sq: number
  avg_supply_block: number
  main_race: string
  total_play_time: number
  improvements?: Record<string, string>
  regressions?: Record<string, string>
  focus_suggestion: string
  strengths?: string[]
  weaknesses?: string[]
  generated_at: string
}

export interface MentorDashboard {
  user: User
  today_stats: DailyProgress | null
  week_stats: WeekStats | null
  active_goals: Goal[]
  recent_games: RecentGame[]
  current_focus: CoachingFocus | null
  weekly_report: WeeklyReport | null
  progress_trend: DailyProgress[]
}

// ============== Auth API ==============

export async function register(email: string, password: string, sc2PlayerName: string): Promise<AuthResponse> {
  const response = await api.post('/auth/register', {
    email,
    password,
    sc2_player_name: sc2PlayerName,
  })
  return response.data
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const response = await api.post('/auth/login', { email, password })
  return response.data
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout')
}

export async function getMe(): Promise<User> {
  const response = await api.get('/auth/me')
  return response.data
}

// ============== Mentor API ==============

export async function getMentorDashboard(): Promise<MentorDashboard> {
  const response = await api.get('/mentor/dashboard')
  return response.data
}

export async function getGoals(): Promise<{ goals: Goal[]; templates: GoalTemplate[] }> {
  const response = await api.get('/mentor/goals')
  return response.data
}

export async function createGoal(
  goalType: string,
  metricName: string,
  targetValue: number,
  comparison?: string
): Promise<Goal> {
  const response = await api.post('/mentor/goals', {
    goal_type: goalType,
    metric_name: metricName,
    target_value: targetValue,
    comparison,
  })
  return response.data
}

export async function deleteGoal(goalId: number): Promise<void> {
  await api.delete(`/mentor/goals/${goalId}`)
}

export async function getProgress(days = 14): Promise<{ progress: DailyProgress[]; days: number }> {
  const response = await api.get('/mentor/progress', { params: { days } })
  return response.data
}

export async function getWeeklyReport(generate = false): Promise<WeeklyReport> {
  const response = await api.get('/mentor/weekly-report', { params: { generate } })
  return response.data
}

export async function setCoachingFocus(focusArea: string, description: string): Promise<CoachingFocus> {
  const response = await api.post('/mentor/focus', { focus_area: focusArea, description })
  return response.data
}

export async function getGoalTemplates(): Promise<GoalTemplate[]> {
  const response = await api.get('/mentor/goal-templates')
  return response.data
}

export default api
