import { useState, useEffect } from 'react';
import { fetcher } from '@/components/Amis/fetcher';

import { MenuItem } from '@/types/menu';

interface UserRoleResponse {
    role: string;
    cluster: string;
    menu_data: MenuItem[] | string;
}

export const useUserRole = () => {
    const [userRole, setUserRole] = useState<string>('');
    const [menuData, setMenuData] = useState<MenuItem[] | string>([]);

    useEffect(() => {
        const fetchUserRole = async () => {
            try {
                const response = await fetcher({
                    url: '/params/user/role',
                    method: 'get'
                });

                if (response.data && typeof response.data === 'object') {
                    const role = response.data.data as UserRoleResponse;
                    setUserRole(role.role);
                    setMenuData(role.menu_data);

                    const originCluster = localStorage.getItem('cluster') || '';
                    if (originCluster === "" && role.cluster !== "") {
                        localStorage.setItem('cluster', role.cluster);
                    }
                }
            } catch (error) {
                console.error('Failed to fetch user role:', error);
            }
        };

        fetchUserRole();
    }, []);

    return { userRole, menuData };
};