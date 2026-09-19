const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';
export function getToken(){ return localStorage.getItem('pulse_token'); }
export function setToken(token){ localStorage.setItem('pulse_token', token); }
export function clearToken(){ localStorage.removeItem('pulse_token'); }
export async function api(path, options={}){
  const headers = new Headers(options.headers || {});
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type','application/json');
  const token = getToken(); if (token) headers.set('Authorization',`Bearer ${token}`);
  const response = await fetch(`${API_URL}${path}`, {...options, headers});
  if (response.status === 204) return null;
  let data = null; try { data = await response.json(); } catch {}
  if (!response.ok) throw new Error(data?.error || 'Something went wrong.');
  return data;
}
export { API_URL };
