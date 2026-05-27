import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { AgentTracesPage } from './agent-trace/AgentTracesPage';
import { getRoutePrefix } from './routePrefix';
import './App.css';

const routePrefix = getRoutePrefix();

function App() {
  return (
    <BrowserRouter basename={routePrefix || undefined}>
      <Routes>
        <Route path="/" element={<AgentTracesPage routeBase="" />} />
        <Route path="/:traceId" element={<AgentTracesPage routeBase="" />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
