import { useState } from 'react';
import { usersApi } from '../../../shared/api/users';
import { userApi } from '../../../entity/users';
import { ToastHelper } from '../../../shared/toast/ToastHelper';

interface ChangePasswordComponentProps {
  onClose?: () => void;
}

export function ChangePasswordComponent({ onClose }: ChangePasswordComponentProps) {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const validateForm = (): string | null => {
    if (!currentPassword) {
      return 'La contraseña actual es requerida';
    }

    if (!newPassword) {
      return 'La nueva contraseña es requerida';
    }

    if (newPassword.length < 8) {
      return 'La nueva contraseña debe tener al menos 8 caracteres';
    }

    if (newPassword === currentPassword) {
      return 'La nueva contraseña debe ser diferente a la actual';
    }

    if (newPassword !== confirmPassword) {
      return 'Las contraseñas no coinciden';
    }

    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const validationError = validateForm();
    if (validationError) {
      ToastHelper.showToast({
        title: 'Error',
        description: validationError,
      });
      return;
    }

    setIsLoading(true);

    try {
      await usersApi.changeMyPassword(currentPassword, newPassword);
      ToastHelper.showToast({
        title: 'Éxito',
        description: 'Contraseña cambiada exitosamente. Cerrando sesión...',
      });
      
      // Reset form
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      
      if (onClose) {
        onClose();
      }

      // Logout after a short delay to show the success message
      setTimeout(() => {
        userApi.logout();
        userApi.notifyAuthListeners();
      }, 1500);
    } catch (error) {
      console.error('Error changing password:', error);
      ToastHelper.showToast({
        title: 'Error',
        description: 'Error al cambiar la contraseña. Verifica que tu contraseña actual sea correcta.',
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-md">
      <h2 className="mb-6 text-2xl font-bold text-gray-800">Cambiar Contraseña</h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="currentPassword" className="mb-2 block text-sm font-medium text-gray-700">
            Contraseña Actual
          </label>
          <input
            id="currentPassword"
            type="password"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200"
            placeholder="Ingresa tu contraseña actual"
            disabled={isLoading}
          />
        </div>

        <div>
          <label htmlFor="newPassword" className="mb-2 block text-sm font-medium text-gray-700">
            Nueva Contraseña
          </label>
          <input
            id="newPassword"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200"
            placeholder="Mínimo 8 caracteres"
            disabled={isLoading}
          />
        </div>

        <div>
          <label htmlFor="confirmPassword" className="mb-2 block text-sm font-medium text-gray-700">
            Confirmar Nueva Contraseña
          </label>
          <input
            id="confirmPassword"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200"
            placeholder="Confirma tu nueva contraseña"
            disabled={isLoading}
          />
        </div>

        <div className="flex gap-3 pt-4">
          <button
            type="submit"
            disabled={isLoading}
            className="flex-1 rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isLoading ? 'Cambiando...' : 'Cambiar Contraseña'}
          </button>
          
          {onClose && (
            <button
              type="button"
              onClick={onClose}
              disabled={isLoading}
              className="flex-1 rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Cancelar
            </button>
          )}
        </div>
      </form>
    </div>
  );
}
