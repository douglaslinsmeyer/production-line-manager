import { Link } from 'react-router-dom';
import { PlusIcon } from '@heroicons/react/24/outline';
import Button from '@/components/common/Button';
import LineList from '@/components/lines/LineList';

export default function Dashboard() {
  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-gray-600">
            Manage and monitor all production lines
          </p>
        </div>
        <Link to="/lines/new">
          <Button variant="primary">
            <PlusIcon className="h-5 w-5 mr-2" />
            Create New Line
          </Button>
        </Link>
      </div>

      {/* Lines list */}
      <LineList />
    </div>
  );
}
