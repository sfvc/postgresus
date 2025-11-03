import React, { useEffect, useState } from 'react';
import { Button, Modal, Spin, Input, Table, Space } from 'antd';
import { usersApi, type UserResponse } from '../../../shared/api/users';
import { ToastHelper } from '../../../shared/toast/ToastHelper';

interface Props {
  contentHeight: number;
}

export const AdminUsersComponent: React.FC<Props> = ({ contentHeight }) => {
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [isShowAddUser, setIsShowAddUser] = useState(false);

  const [newEmail, setNewEmail] = useState('');
  const [newPassword, setNewPassword] = useState('');

  const load = async () => {
    setLoading(true);
    try {
      const list = await usersApi.listUsers();
      setUsers(list);
    } catch (e: any) {
      ToastHelper.showToast({
        title: 'Error',
        description: e.message || String(e)
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleCreate = async () => {
    if (!newEmail || !newPassword) {
      ToastHelper.showToast({
        title: 'Validation Error',
        description: 'Email and password are required'
      });
      return;
    }

    try {
      await usersApi.createUser({ email: newEmail, password: newPassword, role: 'MANAGER' });
      setNewEmail('');
      setNewPassword('');
      setIsShowAddUser(false);
      await load();
      ToastHelper.showToast({
        title: 'Success',
        description: 'User created successfully'
      });
    } catch (e: any) {
      ToastHelper.showToast({
        title: 'Error',
        description: e.message || String(e)
      });
    }
  };

  const toggleStatus = async (user: UserResponse) => {
    const newStatus = user.status === 'ACTIVE' ? 'BLOCKED' : 'ACTIVE';
    try {
      await usersApi.updateUserStatus(user.id, newStatus);
      await load();
      ToastHelper.showToast({
        title: 'Success',
        description: `User ${newStatus === 'ACTIVE' ? 'unblocked' : 'blocked'} successfully`
      });
    } catch (e: any) {
      ToastHelper.showToast({
        title: 'Error',
        description: e.message || String(e)
      });
    }
  };

  const changePassword = async (user: UserResponse) => {
    const pwd = prompt('Enter new password for ' + user.email);
    if (!pwd) return;
    try {
      await usersApi.changeUserPassword(user.id, pwd);
      ToastHelper.showToast({
        title: 'Success',
        description: 'Password changed successfully'
      });
    } catch (e: any) {
      ToastHelper.showToast({
        title: 'Error',
        description: e.message || String(e)
      });
    }
  };

  const columns = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (record: UserResponse) => (
        <Space size="small">
          <Button
            size="small"
            type={record.status === 'ACTIVE' ? 'default' : 'primary'}
            onClick={() => toggleStatus(record)}
          >
            {record.status === 'ACTIVE' ? 'Block' : 'Unblock'}
          </Button>
          <Button
            size="small"
            onClick={() => changePassword(record)}
          >
            Change password
          </Button>
        </Space>
      ),
    },
  ];

  if (loading) {
    return (
      <div className="mx-3 my-3 flex w-[250px] justify-center">
        <Spin />
      </div>
    );
  }

  const addUserButton = (
    <Button type="primary" className="mb-4" onClick={() => setIsShowAddUser(true)}>
      Add User
    </Button>
  );

  return (
    <>
      <div className="flex grow">
        <div
          className="mx-3 w-full overflow-y-auto rounded bg-white p-4 shadow"
          style={{ height: contentHeight }}
        >
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-bold">Admin — Users</h2>
            {addUserButton}
          </div>

          <Table
            dataSource={users}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="small"
          />
        </div>
      </div>

      {isShowAddUser && (
        <Modal
          title="Add User"
          open={isShowAddUser}
          onCancel={() => setIsShowAddUser(false)}
          footer={[
            <Button key="cancel" onClick={() => setIsShowAddUser(false)}>
              Cancel
            </Button>,
            <Button key="create" type="primary" onClick={handleCreate}>
              Create Manager
            </Button>,
          ]}
        >
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Email</label>
              <Input
                placeholder="Email"
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Password</label>
              <Input.Password
                placeholder="Password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
            </div>
          </div>
        </Modal>
      )}
    </>
  );
};

export default AdminUsersComponent;
