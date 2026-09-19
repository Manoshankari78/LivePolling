import React from 'react';
import {createContext,useContext,useEffect,useState} from 'react';
import {api,clearToken,getToken,setToken} from '../services/api';
const AuthContext=createContext(null);
export function AuthProvider({children}){
 const [user,setUser]=useState(null),[loading,setLoading]=useState(true);
 useEffect(()=>{if(!getToken()){setLoading(false);return} api('/auth/me').then(d=>setUser(d.user)).catch(()=>clearToken()).finally(()=>setLoading(false))},[]);
 const login=async(email,password)=>{const d=await api('/auth/login',{method:'POST',body:JSON.stringify({email,password})});setToken(d.token);setUser(d.user)};
 const register=async(payload)=>{const d=await api('/auth/register',{method:'POST',body:JSON.stringify(payload)});setToken(d.token);setUser(d.user)};
 const logout=()=>{clearToken();setUser(null)};
 return <AuthContext.Provider value={{user,loading,login,register,logout}}>{children}</AuthContext.Provider>
}
export const useAuth=()=>useContext(AuthContext);
