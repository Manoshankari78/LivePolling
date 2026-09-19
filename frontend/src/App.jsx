import React from 'react';
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate
} from 'react-router-dom';

import { AuthProvider, useAuth } from './context/AuthContext';
import AppLayout from './layouts/AppLayout';
import Home from './pages/Home';
import { Login, Register } from './pages/Auth';
import Dashboard from './pages/Dashboard';
import CreatePoll from './pages/CreatePoll';
import PollPage from './pages/PollPage';
import Loading from './components/Loading';

function Protected({ children }) {
  const { user, loading } = useAuth();

  if (loading) {
    return <Loading />;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route path="/" element={<Home />} />

            <Route path="/login" element={<Login />} />

            <Route path="/register" element={<Register />} />

            <Route path="/poll/:id" element={<PollPage />} />

            <Route
              path="/dashboard"
              element={
                <Protected>
                  <Dashboard />
                </Protected>
              }
            />

            <Route
              path="/create"
              element={
                <Protected>
                  <CreatePoll />
                </Protected>
              }
            />

            <Route
              path="/poll/:id/edit"
              element={
                <Protected>
                  <CreatePoll edit />
                </Protected>
              }
            />

            <Route
              path="*"
              element={<Navigate to="/" replace />}
            />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
