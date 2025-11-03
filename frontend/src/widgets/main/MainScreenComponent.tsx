import { Tooltip, Button } from 'antd';
import { useEffect, useState } from 'react';
import GitHubButton from 'react-github-btn';

import { APP_VERSION, getApplicationServer } from '../../constants';
import { type DiskUsage, diskApi } from '../../entity/disk';
import { userApi } from '../../entity/users';
import { usersApi, type UserResponse } from '../../shared/api/users';
import { DatabasesComponent } from '../../features/databases/ui/DatabasesComponent';
import { NotifiersComponent } from '../../features/notifiers/ui/NotifiersComponent';
import { StoragesComponent } from '../../features/storages/StoragesComponent';
import { AdminUsersComponent } from '../../features/users';
import { useScreenHeight } from '../../shared/hooks';
import { ToastHelper } from '../../shared/toast/ToastHelper';

export const MainScreenComponent = () => {
  const screenHeight = useScreenHeight();
  const contentHeight = screenHeight - 95;

  const [selectedTab, setSelectedTab] = useState<'notifiers' | 'storages' | 'databases' | 'users'>(
    'databases',
  );
  const [diskUsage, setDiskUsage] = useState<DiskUsage | undefined>(undefined);
  const [currentUser, setCurrentUser] = useState<UserResponse | undefined>(undefined);

  // cargar variables de entorno
  const showCommunityLinks = import.meta.env.VITE_APP_ENVIRONMENT !== 'dev';

  useEffect(() => {
    diskApi
      .getDiskUsage()
      .then((diskUsage) => {
        setDiskUsage(diskUsage);
      })
      .catch((error) => {
        ToastHelper.showToast({
          title: 'Error',
          description: error.message
        });
      });

    // Load current user info - try to get from API first, fallback to localStorage
    usersApi
      .getCurrentUser()
      .then((user) => {
        setCurrentUser(user);
        // Set default tab based on user role
        if (user.role === 'ADMIN') {
          setSelectedTab('users');
          console.log('User is ADMIN, defaulting to Users tab');
        } else if (user.role === 'MANAGER') {
          setSelectedTab('databases');
          console.log('User is MANAGER, defaulting to Databases tab');
        }
      })
      .catch(() => {
        // Fallback: create a mock user object for demo purposes
        // In a real app, you might want to decode JWT token or call a different endpoint
        const mockUser: UserResponse = {
          id: 'current-user',
          email: 'admin@postgresus.com',
          role: 'ADMIN', // Default to ADMIN for now
          status: 'ACTIVE',
          createdAt: new Date().toISOString()
        };

        setCurrentUser(mockUser);
        console.warn('Failed to load user from API, using fallback user info.');

        ToastHelper.showToast({
          title: 'Info',
          description: 'Using fallback user info. Please contact admin if this persists.'
        });
      });
  }, []);
  

  const handleLogout = () => {
    userApi.logout();
    userApi.notifyAuthListeners();
  };

  const isUsedMoreThan95Percent =
    diskUsage && diskUsage.usedSpaceBytes / diskUsage.totalSpaceBytes > 0.95;

  // Define menu items based on user role
  const getMenuItems = () => {
    const allItems = [
      {
        text: 'Databases',
        name: 'databases',
        icon: '/icons/menu/database-gray.svg',
        selectedIcon: '/icons/menu/database-white.svg',
        onClick: () => setSelectedTab('databases'),
        roles: ['MANAGER'],
      },
      {
        text: 'Storages',
        name: 'storages',
        icon: '/icons/menu/storage-gray.svg',
        selectedIcon: '/icons/menu/storage-white.svg',
        onClick: () => setSelectedTab('storages'),
        roles: ['MANAGER'],
      },
      {
        text: 'Notifiers',
        name: 'notifiers',
        icon: '/icons/menu/notifier-gray.svg',
        selectedIcon: '/icons/menu/notifier-white.svg',
        onClick: () => setSelectedTab('notifiers'),
        roles: ['MANAGER'],
      },
      {
        text: 'Users',
        name: 'users',
        icon: '/icons/pen-gray.svg',
        selectedIcon: '/icons/pen-gray.svg',
        onClick: () => setSelectedTab('users'),
        roles: ['ADMIN'],
      },
    ];

    // Filter items based on user role
    return allItems.filter(item => 
      currentUser?.role && item.roles.includes(currentUser.role)
    );
  };

  return (
    <div style={{ height: screenHeight }} className="bg-[#f5f5f5] p-3">
      {/* ===================== NAVBAR ===================== */}
      <div className="mb-3 flex h-[60px] items-center rounded bg-white p-3 shadow">
        <div className="flex items-center gap-3 hover:opacity-80">
          <a href="https://postgresus.com" target="_blank" rel="noreferrer">
            <img className="h-[35px] w-[35px]" src="/logo.svg" />
          </a>

          <div className="text-xl font-bold">
            <a
              href="https://postgresus.com"
              className="text-black"
              target="_blank"
              rel="noreferrer"
            >
              Postgresus
            </a>
          </div>
        </div>

        <div className="mr-3 ml-auto flex items-center gap-5">
          <a
            className="hover:opacity-80"
            href={`${getApplicationServer()}/api/v1/system/health`}
            target="_blank"
            rel="noreferrer"
          >
            Health-check
          </a>

          {showCommunityLinks && (
            <>
              <a
                className="hover:opacity-80"
                href="https://t.me/postgresus_community"
                target="_blank"
                rel="noreferrer"
              >
                Community
              </a>

              <div className="mt-1">
                <GitHubButton
                  href="https://github.com/RostislavDugin/postgresus"
                  data-icon="octicon-star"
                  data-size="large"
                  data-show-count="true"
                  aria-label="Star RostislavDugin/postgresus on GitHub"
                >
                  &nbsp;Star on GitHub
                </GitHubButton>
              </div>
            </>
          )}

          {currentUser && (
            <div className="text-sm text-gray-600">
              {currentUser.email} ({currentUser.role})
            </div>
          )}

          <Button type="primary" size="small" onClick={handleLogout}>
            Logout
          </Button>

          {diskUsage && (
            <Tooltip title="To make backups locally and restore them, you need to have enough space on your disk. For restore, you need to have same amount of space that the backup size.">
              <div
                className={`cursor-pointer text-center text-xs ${isUsedMoreThan95Percent ? 'text-red-500' : 'text-gray-500'}`}
              >
                {(diskUsage.usedSpaceBytes / 1024 ** 3).toFixed(1)} of{' '}
                {(diskUsage.totalSpaceBytes / 1024 ** 3).toFixed(1)} GB
                <br />
                ROM used (
                {((diskUsage.usedSpaceBytes / diskUsage.totalSpaceBytes) * 100).toFixed(1)}%)
              </div>
            </Tooltip>
          )}
        </div>
      </div>
      {/* ===================== END NAVBAR ===================== */}

      <div className="relative flex">
        <div
          className="max-w-[60px] min-w-[60px] rounded bg-white py-2 shadow"
          style={{ height: contentHeight }}
        >
          {getMenuItems().map((tab) => (
            <div key={tab.text} className="flex justify-center">
              <div
                className={`flex h-[50px] w-[50px] cursor-pointer items-center justify-center rounded ${selectedTab === tab.name ? 'bg-blue-600' : 'hover:bg-blue-50'}`}
                onClick={tab.onClick}
              >
                <div className="mb-1">
                  <div className="flex justify-center">
                    <img
                      src={selectedTab === tab.name ? tab.selectedIcon : tab.icon}
                      width={20}
                      alt={tab.text}
                      loading="lazy"
                    />
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {selectedTab === 'notifiers' && <NotifiersComponent contentHeight={contentHeight} />}
        {selectedTab === 'storages' && <StoragesComponent contentHeight={contentHeight} />}
        {selectedTab === 'databases' && <DatabasesComponent contentHeight={contentHeight} />}
        {selectedTab === 'users' && currentUser?.role === 'ADMIN' && <AdminUsersComponent contentHeight={contentHeight} />}

        <div className="absolute bottom-1 left-2 mb-[0px] text-sm text-gray-400">
          v{APP_VERSION}
        </div>
      </div>
    </div>
  );
};
