import { create } from "zustand";
import { api, clearToken, getToken, setToken } from "@/services/api";

interface AuthState {
  token: string | null;
  email: string | null;
  role: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const useAuth = create<AuthState>((set) => ({
  token: getToken(),
  email: null,
  role: null,
  login: async (email, password) => {
    const result = await api.login(email, password);
    setToken(result.access_token);
    set({ token: result.access_token, email: result.email, role: result.role });
  },
  logout: () => {
    clearToken();
    set({ token: null, email: null, role: null });
  },
}));
