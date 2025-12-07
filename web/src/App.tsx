import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/common/Layout';
import Dashboard from './pages/Dashboard';
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
