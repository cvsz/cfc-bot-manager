import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import type { 
  LoginRequest, 
  LoginResponse, 
  RefreshTokenResponse,
  ListResponse,
  PaginationParams,
  Room,
  Bot,
} from '@/types/api';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

class ApiClient {
  private client: AxiosInstance;
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private isRefreshing = false;
  private failedRequestsQueue: Array<{
    resolve: (value: unknown) => void;
    reject: (reason: unknown) => void;
  }> = [];

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: true,
    });

    this.setupInterceptors();
    this.loadTokensFromStorage();
  }

  private setupInterceptors(): void {
    // Request interceptor to add auth header
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        if (this.accessToken && config.headers) {
          config.headers.Authorization = `Bearer ${this.accessToken}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor for token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.isRefreshing) {
            return new Promise((resolve, reject) => {
              this.failedRequestsQueue.push({ resolve, reject });
            }).then(() => this.client(originalRequest));
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            await this.refreshAccessToken();
            this.processQueue(null);
            return this.client(originalRequest);
          } catch (refreshError) {
            this.processQueue(refreshError);
            this.clearAuth();
            window.location.href = '/login';
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        return Promise.reject(error);
      }
    );
  }

  private processQueue(error: unknown): void {
    this.failedRequestsQueue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve(null);
      }
    });
    this.failedRequestsQueue = [];
  }

  private loadTokensFromStorage(): void {
    this.accessToken = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
  }

  private saveTokensToStorage(accessToken: string, refreshToken: string): void {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
  }

  private clearAuth(): void {
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }

  // Auth methods
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.post<LoginResponse>('/auth/login', credentials);
    this.saveTokensToStorage(response.data.access_token, response.data.refresh_token);
    return response.data;
  }

  async logout(): Promise<void> {
    try {
      await this.client.post('/auth/logout');
    } finally {
      this.clearAuth();
    }
  }

  async refreshAccessToken(): Promise<void> {
    if (!this.refreshToken) throw new Error('No refresh token available');
    
    const response = await this.client.post<RefreshTokenResponse>('/auth/refresh', {
      refresh_token: this.refreshToken,
    });
    this.saveTokensToStorage(response.data.access_token, response.data.refresh_token);
  }

  async getCurrentUser(): Promise<LoginResponse['user']> {
    const response = await this.client.get<LoginResponse['user']>('/auth/me');
    return response.data;
  }

  // Generic CRUD methods
  async get<T>(url: string, params?: PaginationParams): Promise<T> {
    const response = await this.client.get<T>(url, { params });
    return response.data;
  }

  async getList<T>(url: string, params?: PaginationParams): Promise<ListResponse<T>> {
    const response = await this.client.get<ListResponse<T>>(url, { params });
    return response.data;
  }

  async post<T>(url: string, data: unknown): Promise<T> {
    const response = await this.client.post<T>(url, data);
    return response.data;
  }

  async put<T>(url: string, data: unknown): Promise<T> {
    const response = await this.client.put<T>(url, data);
    return response.data;
  }

  async patch<T>(url: string, data: unknown): Promise<T> {
    const response = await this.client.patch<T>(url, data);
    return response.data;
  }

  async delete<T>(url: string): Promise<T> {
    const response = await this.client.delete<T>(url);
    return response.data;
  }

  // Organization methods
  async getOrganizations(params?: PaginationParams) {
    return this.getList('/organizations', params);
  }

  async getOrganization(id: string) {
    return this.get(`/organizations/${id}`);
  }

  async createOrganization(data: { name: string; slug: string; description?: string }) {
    return this.post('/organizations', data);
  }

  async updateOrganization(id: string, data: Partial<{ name: string; slug: string; description: string }>) {
    return this.patch(`/organizations/${id}`, data);
  }

  async deleteOrganization(id: string) {
    return this.delete(`/organizations/${id}`);
  }

  // User methods
  async getUsers(organizationId: string, params?: PaginationParams) {
    return this.getList(`/organizations/${organizationId}/users`, params);
  }

  async getUser(organizationId: string, userId: string) {
    return this.get(`/organizations/${organizationId}/users/${userId}`);
  }

  async createUser(organizationId: string, data: { email: string; display_name?: string; password: string; role_ids?: string[] }) {
    return this.post(`/organizations/${organizationId}/users`, data);
  }

  async updateUser(organizationId: string, userId: string, data: Partial<{ email: string; display_name: string; is_active: boolean; role_ids: string[] }>) {
    return this.patch(`/organizations/${organizationId}/users/${userId}`, data);
  }

  async deleteUser(organizationId: string, userId: string) {
    return this.delete(`/organizations/${organizationId}/users/${userId}`);
  }

  // Room methods
  async getRooms(organizationId: string, params?: PaginationParams) {
    return this.getList<Room>(`/organizations/${organizationId}/rooms`, params);
  }

  async getRoom(organizationId: string, roomId: string) {
    return this.get(`/organizations/${organizationId}/rooms/${roomId}`);
  }

  async createRoom(organizationId: string, data: { name: string; access_key: string; language?: string; config?: Partial<import('@/types/api').RoomConfig> }) {
    return this.post(`/organizations/${organizationId}/rooms`, data);
  }

  async updateRoom(organizationId: string, roomId: string, data: Partial<{ name: string; access_key: string; language: string; config: import('@/types/api').RoomConfig; bot_id: string | null }>) {
    return this.patch(`/organizations/${organizationId}/rooms/${roomId}`, data);
  }

  async deleteRoom(organizationId: string, roomId: string) {
    return this.delete(`/organizations/${organizationId}/rooms/${roomId}`);
  }

  // Bot methods
  async getBots(organizationId: string, params?: PaginationParams) {
    return this.getList<Bot>(`/organizations/${organizationId}/bots`, params);
  }

  async getBot(organizationId: string, botId: string) {
    return this.get(`/organizations/${organizationId}/bots/${botId}`);
  }

  async createBot(organizationId: string, data: { name: string; username: string; password: string; room_id?: string; config?: Partial<import('@/types/api').BotConfig> }) {
    return this.post(`/organizations/${organizationId}/bots`, data);
  }

  async updateBot(organizationId: string, botId: string, data: Partial<{ name: string; username: string; password: string; room_id: string | null; config: import('@/types/api').BotConfig }>) {
    return this.patch(`/organizations/${organizationId}/bots/${botId}`, data);
  }

  async deleteBot(organizationId: string, botId: string) {
    return this.delete(`/organizations/${organizationId}/bots/${botId}`);
  }

  // Camfrog Adapter methods (bot/room/plugin operations via adapter)
  async getBotStatus(organizationId: string, botId: string) {
    return this.get(`/organizations/${organizationId}/bots/${botId}/status`);
  }

  async setBotLanguage(organizationId: string, botId: string, locale: string) {
    return this.put(`/organizations/${organizationId}/bots/${botId}/language`, { locale });
  }

  async toggleBotPlugin(organizationId: string, botId: string, pluginId: number, enabled: boolean) {
    return this.put(`/organizations/${organizationId}/bots/${botId}/plugins/${pluginId}`, { enabled });
  }

  async startBot(organizationId: string, botId: string) {
    return this.post(`/organizations/${organizationId}/bots/${botId}/start`, {});
  }

  async stopBot(organizationId: string, botId: string) {
    return this.post(`/organizations/${organizationId}/bots/${botId}/stop`, {});
  }

  async restartBot(organizationId: string, botId: string) {
    return this.post(`/organizations/${organizationId}/bots/${botId}/restart`, {});
  }

  async pauseRoom(organizationId: string, roomId: string, pause: boolean) {
    return this.post(`/organizations/${organizationId}/rooms/${roomId}/pause`, { pause });
  }

  async setRoomLanguage(organizationId: string, roomId: string, locale: string) {
    return this.put(`/organizations/${organizationId}/rooms/${roomId}/language`, { locale });
  }

  async getPluginSettings(organizationId: string, pluginId: number, roomId: string) {
    return this.get(`/organizations/${organizationId}/plugins/${pluginId}/settings`, { room_id: roomId } as unknown as PaginationParams);
  }

  async savePluginSettings(organizationId: string, pluginId: number, roomId: string, settings: Record<string, unknown>) {
    return this.put(`/organizations/${organizationId}/plugins/${pluginId}/settings`, { ...settings, room_id: roomId } as unknown as PaginationParams);
  }

  // Plugin methods
  async getPlugins(params?: PaginationParams) {
    return this.getList('/plugins', params);
  }

  async getPlugin(pluginId: string) {
    return this.get(`/plugins/${pluginId}`);
  }

  async getPluginConfigs(organizationId: string, roomId?: string, botId?: string, params?: PaginationParams) {
    const queryParams = new URLSearchParams();
    if (roomId) queryParams.append('room_id', roomId);
    if (botId) queryParams.append('bot_id', botId);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) queryParams.append(key, String(value));
      });
    }
    return this.getList(`/organizations/${organizationId}/plugin-configs`, queryParams as unknown as PaginationParams);
  }

  async createPluginConfig(organizationId: string, data: { plugin_id: string; room_id?: string; bot_id?: string; settings: Record<string, unknown>; is_enabled?: boolean }) {
    return this.post(`/organizations/${organizationId}/plugin-configs`, data);
  }

  async updatePluginConfig(organizationId: string, configId: string, data: { settings?: Record<string, unknown>; is_enabled?: boolean; version: number }) {
    return this.patch(`/organizations/${organizationId}/plugin-configs/${configId}`, data);
  }

  async deletePluginConfig(organizationId: string, configId: string) {
    return this.delete(`/organizations/${organizationId}/plugin-configs/${configId}`);
  }

  // Audit log methods
  async getAuditLogs(organizationId: string, params?: PaginationParams & { resource_type?: string; resource_id?: string; actor_user_id?: string }) {
    return this.getList(`/organizations/${organizationId}/audit-logs`, params);
  }

  // API Key methods
  async getAPIKeys(organizationId: string, params?: PaginationParams) {
    return this.getList(`/organizations/${organizationId}/api-keys`, params);
  }

  async createAPIKey(organizationId: string, data: { name: string; permissions: string[]; expires_at?: string }) {
    return this.post(`/organizations/${organizationId}/api-keys`, data);
  }

  async revokeAPIKey(organizationId: string, keyId: string) {
    return this.post(`/organizations/${organizationId}/api-keys/${keyId}/revoke`, {});
  }

  async deleteAPIKey(organizationId: string, keyId: string) {
    return this.delete(`/organizations/${organizationId}/api-keys/${keyId}`);
  }

  // Bot Connection methods
  async getBotConnections(organizationId: string, botId: string, params?: PaginationParams) {
    return this.getList(`/organizations/${organizationId}/bots/${botId}/connections`, params);
  }

  // Role methods
  async getRoles(organizationId: string, params?: PaginationParams) {
    return this.getList(`/organizations/${organizationId}/roles`, params);
  }

  async getRole(organizationId: string, roleId: string) {
    return this.get(`/organizations/${organizationId}/roles/${roleId}`);
  }

  async createRole(organizationId: string, data: { name: string; permission_ids: string[] }) {
    return this.post(`/organizations/${organizationId}/roles`, data);
  }

  async updateRole(organizationId: string, roleId: string, data: { name?: string; permission_ids?: string[] }) {
    return this.patch(`/organizations/${organizationId}/roles/${roleId}`, data);
  }

  async deleteRole(organizationId: string, roleId: string) {
    return this.delete(`/organizations/${organizationId}/roles/${roleId}`);
  }

  // Permission methods
  async getPermissions() {
    return this.get('/permissions');
  }

  // MFA methods
  async setupMFA() {
    return this.get('/auth/mfa/setup');
  }

  async verifyMFA(code: string) {
    return this.post('/auth/mfa/verify', { code });
  }

  async disableMFA() {
    return this.post('/auth/mfa/disable', {});
  }

  // Check if authenticated
  isAuthenticated(): boolean {
    return !!this.accessToken;
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }
}

export const api = new ApiClient();
export default api;