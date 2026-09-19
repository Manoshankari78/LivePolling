function uuid(){if(crypto?.randomUUID)return crypto.randomUUID();return `${Date.now()}-${Math.random().toString(36).slice(2)}-${Math.random().toString(36).slice(2)}`}
export function getVoterId(){let id=localStorage.getItem('pulse_voter_id');if(!id){id=uuid();localStorage.setItem('pulse_voter_id',id)}return id}
