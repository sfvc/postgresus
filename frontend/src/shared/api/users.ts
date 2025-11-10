import { getApplicationServer } from '../../constants';
import RequestOptions from './RequestOptions';
import { apiHelper } from './apiHelper';

export interface CreateUserRequest {
  email: string;
  password: string;
  role: string;
}

export interface UserResponse {
  id: string;
  email: string;
  role: string;
  status: string;
  createdAt: string;
}

export const usersApi = {
  createUser: async (req: CreateUserRequest): Promise<void> => {
    const requestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify(req));
    await apiHelper.fetchPostRaw(`${getApplicationServer()}/api/v1/users/admin/create-user`, requestOptions);
  },

  listUsers: async (): Promise<UserResponse[]> => {
    const requestOptions = new RequestOptions();
    return apiHelper.fetchGetJson<UserResponse[]>(`${getApplicationServer()}/api/v1/users/admin/list`, requestOptions);
  },

  updateUserStatus: async (userId: string, status: string): Promise<UserResponse> => {
    const requestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify({ status }));
    return apiHelper.fetchPutJson<UserResponse>(`${getApplicationServer()}/api/v1/users/admin/${userId}/status`, requestOptions);
  },

  changeUserPassword: async (userId: string, newPassword: string): Promise<void> => {
    const requestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify({ newPassword }));
    await apiHelper.fetchPutJson<void>(`${getApplicationServer()}/api/v1/users/admin/${userId}/password`, requestOptions);
  },

  getCurrentUser: async (): Promise<UserResponse> => {
    const requestOptions = new RequestOptions();
    return apiHelper.fetchGetJson<UserResponse>(`${getApplicationServer()}/api/v1/users/me`, requestOptions);
  },

  changeMyPassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    const requestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify({ currentPassword, newPassword }));
    await apiHelper.fetchPutJson<void>(`${getApplicationServer()}/api/v1/users/me/password`, requestOptions);
  },
};

export default usersApi;
