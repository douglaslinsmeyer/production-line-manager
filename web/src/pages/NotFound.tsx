import { Link } from 'react-router-dom';
import { HomeIcon } from '@heroicons/react/24/outline';
import Button from '@/components/common/Button';

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <h1 className="text-6xl font-bold text-gray-900 mb-4">404</h1>
      <h2 className="text-2xl font-semibold text-gray-700 mb-2">Page Not Found</h2>
      <p className="text-gray-600 mb-8">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link to="/">
        <Button variant="primary">
          <HomeIcon className="h-5 w-5 mr-2" />
          Back to Dashboard
        </Button>
      </Link>
    </div>
  );
}
