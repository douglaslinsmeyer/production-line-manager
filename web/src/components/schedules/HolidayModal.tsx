import { Fragment } from 'react';
import { Dialog, DialogPanel, DialogTitle, Transition, TransitionChild } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { ScheduleHoliday } from '@/api/types';

const holidaySchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  holiday_date: z.string().min(1, 'Date is required'),
});

type HolidayFormData = z.infer<typeof holidaySchema>;

interface HolidayModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { holiday_date: string; name: string }) => void;
  holiday?: ScheduleHoliday | null;
  isLoading?: boolean;
}

export default function HolidayModal({
  isOpen,
  onClose,
  onSubmit,
  holiday,
  isLoading = false,
}: HolidayModalProps) {
  const isEditing = !!holiday;

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<HolidayFormData>({
    resolver: zodResolver(holidaySchema),
    defaultValues: {
      name: holiday?.name || '',
      holiday_date: holiday?.holiday_date || '',
    },
  });

  const onFormSubmit = (data: HolidayFormData) => {
    onSubmit(data);
    reset();
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Transition show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={handleClose}>
        <TransitionChild
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>

        <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <TransitionChild
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enterTo="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 translate-y-0 sm:scale-100"
              leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <DialogPanel className="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-md sm:p-6">
                <div className="absolute right-0 top-0 pr-4 pt-4">
                  <button
                    type="button"
                    className="rounded-md bg-white text-gray-400 hover:text-gray-500"
                    onClick={handleClose}
                  >
                    <span className="sr-only">Close</span>
                    <XMarkIcon className="h-6 w-6" />
                  </button>
                </div>

                <DialogTitle as="h3" className="text-lg font-semibold text-gray-900 mb-4">
                  {isEditing ? 'Edit Holiday' : 'Add Holiday'}
                </DialogTitle>

                <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
                  <div>
                    <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                      Holiday Name *
                    </label>
                    <input
                      id="name"
                      type="text"
                      {...register('name')}
                      className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="e.g., Christmas Day"
                    />
                    {errors.name && (
                      <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
                    )}
                  </div>

                  <div>
                    <label htmlFor="holiday_date" className="block text-sm font-medium text-gray-700 mb-1">
                      Date *
                    </label>
                    <input
                      id="holiday_date"
                      type="date"
                      {...register('holiday_date')}
                      className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {errors.holiday_date && (
                      <p className="mt-1 text-sm text-red-600">{errors.holiday_date.message}</p>
                    )}
                  </div>

                  <div className="flex justify-end gap-3 pt-4">
                    <button
                      type="button"
                      onClick={handleClose}
                      disabled={isLoading}
                      className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 disabled:opacity-50"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isLoading}
                      className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 disabled:opacity-50"
                    >
                      {isLoading ? 'Saving...' : isEditing ? 'Update' : 'Add'}
                    </button>
                  </div>
                </form>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </Transition>
  );
}
