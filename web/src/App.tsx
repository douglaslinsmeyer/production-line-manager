import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/common/Layout';
import Dashboard from './pages/Dashboard';
import AnalyticsPage from './pages/AnalyticsPage';
import LabelsPage from './pages/LabelsPage';
import LineDetail from './pages/LineDetail';
import CreateLine from './pages/CreateLine';
import EditLine from './pages/EditLine';
import SchedulesPage from './pages/SchedulesPage';
import CreateSchedule from './pages/CreateSchedule';
import ScheduleDetail from './pages/ScheduleDetail';
import EditSchedule from './pages/EditSchedule';
import DeviceDiscovery from './pages/DeviceDiscovery';
import NotFound from './pages/NotFound';

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/analytics" element={<AnalyticsPage />} />
          <Route path="/labels" element={<LabelsPage />} />
          <Route path="/devices" element={<DeviceDiscovery />} />
          <Route path="/lines/new" element={<CreateLine />} />
          <Route path="/lines/:id" element={<LineDetail />} />
          <Route path="/lines/:id/edit" element={<EditLine />} />
          <Route path="/schedules" element={<SchedulesPage />} />
          <Route path="/schedules/new" element={<CreateSchedule />} />
          <Route path="/schedules/:id" element={<ScheduleDetail />} />
          <Route path="/schedules/:id/edit" element={<EditSchedule />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}

export default App;
