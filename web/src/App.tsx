import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/common/Layout';
import Dashboard from './pages/Dashboard';
import AnalyticsPage from './pages/AnalyticsPage';
import LabelsPage from './pages/LabelsPage';
import LineDetail from './pages/LineDetail';
import CreateLine from './pages/CreateLine';
import EditLine from './pages/EditLine';
import NotFound from './pages/NotFound';

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/analytics" element={<AnalyticsPage />} />
          <Route path="/labels" element={<LabelsPage />} />
          <Route path="/lines/new" element={<CreateLine />} />
          <Route path="/lines/:id" element={<LineDetail />} />
          <Route path="/lines/:id/edit" element={<EditLine />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}

export default App;
