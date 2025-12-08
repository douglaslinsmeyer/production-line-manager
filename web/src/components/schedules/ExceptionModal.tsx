import { Fragment, useState } from 'react';
import { Dialog, DialogPanel, DialogTitle, Transition, TransitionChild } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { CreateScheduleExceptionRequest, CreateScheduleDayInput } from '@/api/types';
import WeeklyScheduleEditor from './WeeklyScheduleEditor';
import { createDefaultDays } from './scheduleUtils';

const exceptionSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  start_date: z.string().min(1, 'Start date is required'),
  end_date: z.string().min(1, 'End date is required'),
}).refine((data) => data.start_date <= data.end_date, {
  message: 'End date must be on or after start date',
  path: ['end_date'],
});

type ExceptionFormData = z.infer<typeof exceptionSchema>;

interface ExceptionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateScheduleExceptionRequest) => void;
  isLoading?: boolean;
}

export default function ExceptionModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: ExceptionModalProps) {
  const [days, setDays] = useState<CreateScheduleDayInput[]>(createDefaultDays());

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<ExceptionFormData>({
    resolver: zodResolver(exceptionSchema),
    defaultValues: {
      name: '',
      start_date: '',
      end_date: '',
    },
  });

  const onFormSubmit = (data: ExceptionFormData) => {
    const request: CreateScheduleExceptionRequest = {
      name: data.name,
      start_date: data.start_date,
      end_date: data.end_date,
      days,
    };
    onSubmit(request);
    reset();
    setDays(createDefaultDays());
  };

  const handleClose = () => {
    reset();
    setDays(createDefaultDays());
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
              <DialogPanel className="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-4xl sm:p-6">
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
                  Add Schedule Exception
                </DialogTitle>

                <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-6">
                  {/* Basic Info */}
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="md:col-span-3">
                      <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                        Exception Name *
                      </label>
                      <input
                        id="name"
                        type="text"
                        {...register('name')}
                        className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="e.g., Holiday Season"
                      />
                      {errors.name && (
                        <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="start_date" className="block text-sm font-medium text-gray-700 mb-1">
                        Start Date *
                      </label>
                      <input
                        id="start_date"
                        type="date"
                        {...register('start_date')}
                        className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      {errors.start_date && (
                        <p className="mt-1 text-sm text-red-600">{errors.start_date.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="end_date" className="block text-sm font-medium text-gray-700 mb-1">
                        End Date *
                      </label>
                      <input
                        id="end_date"
                        type="date"
                        {...register('end_date')}
                        className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      {errors.end_date && (
                        <p className="mt-1 text-sm text-red-600">{errors.end_date.message}</p>
                      )}
                    </div>
                  </div>

                  {/* Weekly Schedule Override */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">
                      Override Schedule (for this date range)
                    </h4>
                    <div className="border border-gray-200 rounded-lg overflow-hidden">
                      <WeeklyScheduleEditor days={days} onChange={setDays} disabled={isLoading} />
                    </div>
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
                      {isLoading ? 'Saving...' : 'Add Exception'}
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
