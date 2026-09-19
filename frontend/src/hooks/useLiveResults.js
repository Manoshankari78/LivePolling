import {useEffect,useState} from 'react';
import {api} from '../services/api';
const WS_URL=(import.meta.env.VITE_WS_URL||'ws://localhost:8080/api').replace(/\/$/,'');
export function useLiveResults(pollId){
 const [results,setResults]=useState(null),[connection,setConnection]=useState('connecting');
 useEffect(()=>{let socket,active=true,retry=0,timer;
  const load=()=>api(`/polls/${pollId}/results`).then(d=>active&&setResults(d.results)).catch(()=>{});
  const connect=()=>{if(!active)return;setConnection('connecting');socket=new WebSocket(`${WS_URL}/polls/${pollId}/ws`);socket.onopen=()=>{retry=0;setConnection('live')};socket.onmessage=e=>{try{const data=JSON.parse(e.data);if(data.pollId===pollId)setResults(data)}catch{}};socket.onclose=()=>{if(!active)return;setConnection('reconnecting');timer=setTimeout(connect,Math.min(1000*2**retry,8000));retry++};socket.onerror=()=>setConnection('reconnecting')};
  load();connect();return()=>{active=false;clearTimeout(timer);socket?.close()}
 },[pollId]);
 return {results,connection,refresh:()=>api(`/polls/${pollId}/results`).then(d=>setResults(d.results))};
}
